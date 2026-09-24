package iqlog

// Generic value encoding for Event.Any and the slog adapter.
//
// Two layers sit under appendGenericValue, both producing exactly the bytes
// json.Marshal would:
//
//   - A per-type plan for struct types (and pointers to them) whose exported
//     fields are all primitives. The plan reads each field at its offset and
//     appends it with encoders that mirror encoding/json's number and string
//     formatting, so the reflective walk, the interface boxing per field and
//     the result-slice allocation of json.Marshal all disappear. Every plan is
//     verified against json.Marshal on a probe value before it is published;
//     a type it cannot reproduce exactly is recorded as having no plan.
//
//   - A pooled json.Encoder whose io.Writer appends into the record buffer.
//     Encode hands the sink the complete encoding in a single Write, so the
//     only cost above the encoder's own work is that copy, and no temporary
//     result slice is allocated the way json.Marshal's is.
//
// Anything a plan does not accept -- nested structs, containers, embedded
// fields, tag options, custom marshalers, json.Number -- takes the encoder,
// which is encoding/json itself and therefore identical by construction.

import (
	"encoding/json"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"unicode/utf8"
	"unsafe"
)

// --- pooled encoder ---------------------------------------------------------

// jsonSink is the io.Writer a pooled json.Encoder writes into. Between calls
// dst is nil so the pool never pins a record buffer.
type jsonSink struct {
	dst []byte
}

func (s *jsonSink) Write(p []byte) (int, error) {
	s.dst = append(s.dst, p...)
	return len(p), nil
}

// pooledJSONEncoder keeps json.Marshal's escaping: NewEncoder escapes HTML by
// default exactly as Marshal does, so SetEscapeHTML is deliberately not
// called. The encoder is small (its working buffer is pooled inside
// encoding/json), so sync.Pool's GC-driven bound is enough.
type pooledJSONEncoder struct {
	enc  *json.Encoder
	sink jsonSink
}

var jsonEncoderPool = sync.Pool{New: func() any {
	pe := &pooledJSONEncoder{}
	pe.enc = json.NewEncoder(&pe.sink)
	return pe
}}

// appendMarshalJSON appends json.Marshal(v) to dst without allocating the
// intermediate result. On error nothing has been appended: Encode marshals
// into its own buffer and writes only a complete encoding. A marshal error
// is not sticky on the encoder -- only a Write failure is, and the sink
// never fails -- so the encoder goes back to the pool either way.
func appendMarshalJSON(dst []byte, v any) ([]byte, error) {
	pe := jsonEncoderPool.Get().(*pooledJSONEncoder)
	pe.sink.dst = dst
	err := pe.enc.Encode(v)
	dst = pe.sink.dst
	pe.sink.dst = nil
	jsonEncoderPool.Put(pe)
	if err != nil {
		return dst, err
	}
	// Encode terminates every value with a newline that Marshal does not emit.
	if n := len(dst); n > 0 && dst[n-1] == '\n' {
		dst = dst[:n-1]
	}
	return dst, nil
}

// --- struct plans -----------------------------------------------------------

// maxStructPlans bounds the plan table, counting the types recorded as having
// no plan. Once it is full, new types take the encoder path without touching
// the table, so the hot path never blocks on the table lock.
const maxStructPlans = 512

type planField struct {
	// key is the encoded member name with its trailing colon and, for every
	// field but the first, its leading comma: `,"Name":`.
	key    string
	offset uintptr
	kind   reflect.Kind
}

// structPlan encodes one struct type. A nil *structPlan in the table means
// the type was examined and must use the encoder.
type structPlan struct {
	fields []planField
	// ptr is set when the planned type is *T: the interface data word is
	// then the struct address itself rather than a pointer to a boxed copy.
	ptr bool
}

var (
	// structPlans is an immutable map replaced whole under structPlansMu, so
	// the hot path reads it with one atomic load and no lock.
	structPlans     atomic.Pointer[map[reflect.Type]*structPlan]
	structPlansMu   sync.Mutex
	structPlansFull atomic.Bool
)

// structPlanFor returns the plan for v's dynamic type, building and
// publishing it on first sight, or nil when the type must use the encoder.
func structPlanFor(v any) *structPlan {
	t := reflect.TypeOf(v)
	if t == nil {
		return nil
	}
	switch t.Kind() {
	case reflect.Struct:
	case reflect.Pointer:
		if t.Elem().Kind() != reflect.Struct {
			return nil
		}
	default:
		return nil
	}
	if m := structPlans.Load(); m != nil {
		if p, ok := (*m)[t]; ok {
			return p
		}
	}
	if structPlansFull.Load() {
		return nil
	}
	return publishStructPlan(t)
}

func publishStructPlan(t reflect.Type) *structPlan {
	structPlansMu.Lock()
	defer structPlansMu.Unlock()
	var n int
	old := structPlans.Load()
	if old != nil {
		if p, ok := (*old)[t]; ok {
			return p // another goroutine published it first
		}
		n = len(*old)
	}
	if n >= maxStructPlans {
		structPlansFull.Store(true)
		return nil
	}
	p := compileStructPlan(t)
	m := make(map[reflect.Type]*structPlan, n+1)
	if old != nil {
		for k, v := range *old {
			m[k] = v
		}
	}
	m[t] = p
	structPlans.Store(&m)
	return p
}

// planMarshalMethods are the method names encoding/json consults before its
// default encoding. A type carrying any of them, on T or *T, is left to the
// encoder rather than second-guessing when each applies.
var planMarshalMethods = [...]string{"MarshalJSON", "MarshalJSONTo", "MarshalJSONV2", "MarshalText", "AppendText"}

func hasMarshalMethod(t reflect.Type) bool {
	pt := t
	if t.Kind() != reflect.Pointer {
		pt = reflect.PointerTo(t)
	}
	for _, name := range planMarshalMethods {
		if _, ok := t.MethodByName(name); ok {
			return true
		}
		if _, ok := pt.MethodByName(name); ok {
			return true
		}
	}
	return false
}

var jsonNumberType = reflect.TypeFor[json.Number]()

// planTagName accepts only a bare member name of identifier-like characters:
// no options, no quoting, nothing encoding/json might treat as malformed.
func planTagName(tag string) bool {
	if tag == "" {
		return false
	}
	for i := 0; i < len(tag); i++ {
		c := tag[i]
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

func planKindSupported(k reflect.Kind) bool {
	switch k {
	case reflect.Bool, reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// compileStructPlan builds a plan for t, or returns nil when t has any
// feature whose encoding/json treatment the plan does not reproduce. The
// rules mirror encoding/json's field selection for the cases accepted:
// unexported fields are skipped, `json:"-"` drops a field, a bare tag name
// renames it, and everything else is refused.
func compileStructPlan(t reflect.Type) *structPlan {
	p := &structPlan{}
	st := t
	if t.Kind() == reflect.Pointer {
		p.ptr = true
		st = t.Elem()
	}
	if hasMarshalMethod(st) {
		return nil
	}
	var (
		probeIndex []int    // struct field index of each planned field
		names      []string // member name of each planned field
	)
	for i := 0; i < st.NumField(); i++ {
		sf := st.Field(i)
		if sf.Anonymous {
			return nil // promotion and dominance rules: leave to encoding/json
		}
		if !sf.IsExported() {
			continue
		}
		name := sf.Name
		if tag, ok := sf.Tag.Lookup("json"); ok && tag != "" {
			if tag == "-" {
				continue
			}
			if !planTagName(tag) {
				return nil
			}
			name = tag
		}
		for _, existing := range names {
			if strings.EqualFold(existing, name) {
				return nil // duplicate members follow dominance rules
			}
		}
		if !planKindSupported(sf.Type.Kind()) || sf.Type == jsonNumberType || hasMarshalMethod(sf.Type) {
			return nil
		}
		key, err := json.Marshal(name)
		if err != nil {
			return nil
		}
		if len(p.fields) > 0 {
			key = append([]byte{','}, key...)
		}
		key = append(key, ':')
		p.fields = append(p.fields, planField{key: string(key), offset: sf.Offset, kind: sf.Type.Kind()})
		probeIndex = append(probeIndex, i)
		names = append(names, name)
	}
	if !p.verify(st, probeIndex) {
		return nil
	}
	return p
}

// verify encodes a probe value with distinctive field contents through both
// the plan and json.Marshal and accepts the plan only when the bytes agree.
// This is what makes the plan safe to trust: a field offset, key, tag rule
// or escaping detail the plan got wrong shows up here, once per type, and
// the type is quietly left to the encoder.
func (p *structPlan) verify(st reflect.Type, probeIndex []int) bool {
	probe := reflect.New(st)
	for _, i := range probeIndex {
		f := probe.Elem().Field(i)
		switch f.Kind() {
		case reflect.String:
			// escapes of every class, a JS line separator and an ill-formed byte
			f.SetString("pr\"\\<&>\n\t\u2028\xffobe")
		case reflect.Bool:
			f.SetBool(true)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			f.SetInt(-123)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			f.SetUint(200)
		case reflect.Float32, reflect.Float64:
			f.SetFloat(1.5e-7) // exponent form with the e-07 -> e-7 rewrite
		}
	}
	var value any
	if p.ptr {
		value = probe.Interface()
	} else {
		value = probe.Elem().Interface()
	}
	want, err := json.Marshal(value)
	if err != nil {
		return false
	}
	got, ok := p.appendJSON(nil, value)
	return ok && string(got) == string(want)
}

// anyData returns the interface's data word. For a pointer type it is the
// pointer itself; for a struct type it addresses the copy the interface
// boxed. Plans only accept structs made of non-pointer-shaped fields, so an
// accepted struct is never stored directly in the word.
func anyData(v any) unsafe.Pointer {
	return (*[2]unsafe.Pointer)(unsafe.Pointer(&v))[1]
}

// appendJSON appends the encoding of v, whose dynamic type is the planned
// one. It reports false, having appended nothing, when a float is NaN or
// infinite: encoding/json refuses those, and the caller lets it word the
// refusal.
func (p *structPlan) appendJSON(dst []byte, v any) ([]byte, bool) {
	base := anyData(v)
	if p.ptr && base == nil {
		return append(dst, "null"...), true
	}
	start := len(dst)
	dst = append(dst, '{')
	for i := range p.fields {
		f := &p.fields[i]
		dst = append(dst, f.key...)
		fp := unsafe.Add(base, f.offset)
		switch f.kind {
		case reflect.String:
			dst = appendMarshalString(dst, *(*string)(fp))
		case reflect.Bool:
			if *(*bool)(fp) {
				dst = append(dst, "true"...)
			} else {
				dst = append(dst, "false"...)
			}
		case reflect.Int:
			dst = strconv.AppendInt(dst, int64(*(*int)(fp)), 10)
		case reflect.Int8:
			dst = strconv.AppendInt(dst, int64(*(*int8)(fp)), 10)
		case reflect.Int16:
			dst = strconv.AppendInt(dst, int64(*(*int16)(fp)), 10)
		case reflect.Int32:
			dst = strconv.AppendInt(dst, int64(*(*int32)(fp)), 10)
		case reflect.Int64:
			dst = strconv.AppendInt(dst, *(*int64)(fp), 10)
		case reflect.Uint:
			dst = strconv.AppendUint(dst, uint64(*(*uint)(fp)), 10)
		case reflect.Uint8:
			dst = strconv.AppendUint(dst, uint64(*(*uint8)(fp)), 10)
		case reflect.Uint16:
			dst = strconv.AppendUint(dst, uint64(*(*uint16)(fp)), 10)
		case reflect.Uint32:
			dst = strconv.AppendUint(dst, uint64(*(*uint32)(fp)), 10)
		case reflect.Uint64:
			dst = strconv.AppendUint(dst, *(*uint64)(fp), 10)
		case reflect.Uintptr:
			dst = strconv.AppendUint(dst, uint64(*(*uintptr)(fp)), 10)
		case reflect.Float32:
			f := float64(*(*float32)(fp))
			if math.IsNaN(f) || math.IsInf(f, 0) {
				return dst[:start], false
			}
			dst = appendMarshalFloat(dst, f, 32)
		case reflect.Float64:
			f := *(*float64)(fp)
			if math.IsNaN(f) || math.IsInf(f, 0) {
				return dst[:start], false
			}
			dst = appendMarshalFloat(dst, f, 64)
		}
	}
	return append(dst, '}'), true
}

// finite reports whether every float field of v is a number encoding/json
// would accept. The console encoder uses it to skip the JSON pre-flight that
// otherwise guards fmt against values encoding/json refuses.
func (p *structPlan) finite(v any) bool {
	base := anyData(v)
	if p.ptr && base == nil {
		return true
	}
	for i := range p.fields {
		f := &p.fields[i]
		var x float64
		switch f.kind {
		case reflect.Float32:
			x = float64(*(*float32)(unsafe.Add(base, f.offset)))
		case reflect.Float64:
			x = *(*float64)(unsafe.Add(base, f.offset))
		default:
			continue
		}
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return false
		}
	}
	return true
}

// --- encoding/json-compatible primitives ------------------------------------

// appendMarshalFloat mirrors encoding/json's number formatting: shortest
// representation, 'e' notation outside [1e-6, 1e21), and a two-digit
// negative exponent trimmed of its leading zero.
func appendMarshalFloat(dst []byte, f float64, bits int) []byte {
	abs := math.Abs(f)
	format := byte('f')
	if abs != 0 {
		if bits == 64 && (abs < 1e-6 || abs >= 1e21) || bits == 32 && (float32(abs) < 1e-6 || float32(abs) >= 1e21) {
			format = 'e'
		}
	}
	dst = strconv.AppendFloat(dst, f, format, -1, bits)
	if format == 'e' {
		n := len(dst)
		if n >= 4 && dst[n-4] == 'e' && dst[n-3] == '-' && dst[n-2] == '0' {
			dst[n-2] = dst[n-1]
			dst = dst[:n-1]
		}
	}
	return dst
}

// appendMarshalString appends s as json.Marshal encodes a string, quotes
// included: control characters use their short escapes where JSON has one,
// '<', '>' and '&' are escaped for HTML, U+2028 and U+2029 for JavaScript,
// and each ill-formed byte becomes U+FFFD. It is a separate function from
// appendJSONEscaped, which is the record's own (HTML-agnostic) escaping;
// the two are not interchangeable.
func appendMarshalString(dst []byte, s string) []byte {
	dst = append(dst, '"')
	start := 0
	for i := 0; i < len(s); {
		c := s[i]
		if c < utf8.RuneSelf {
			if c >= 0x20 && c != '"' && c != '\\' && c != '<' && c != '>' && c != '&' {
				i++
				continue
			}
			dst = append(dst, s[start:i]...)
			switch c {
			case '"', '\\':
				dst = append(dst, '\\', c)
			case '\b':
				dst = append(dst, '\\', 'b')
			case '\f':
				dst = append(dst, '\\', 'f')
			case '\n':
				dst = append(dst, '\\', 'n')
			case '\r':
				dst = append(dst, '\\', 'r')
			case '\t':
				dst = append(dst, '\\', 't')
			default:
				dst = append(dst, '\\', 'u', '0', '0', hex[c>>4], hex[c&0xf])
			}
			i++
			start = i
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		switch {
		case r == utf8.RuneError && size == 1:
			dst = append(dst, s[start:i]...)
			dst = append(dst, "\ufffd"...)
		case r == '\u2028' || r == '\u2029':
			dst = append(dst, s[start:i]...)
			dst = append(dst, '\\', 'u', '2', '0', '2', hex[r&0xf])
		default:
			i += size
			continue
		}
		i += size
		start = i
	}
	dst = append(dst, s[start:]...)
	return append(dst, '"')
}
