package iqlog

// Differential coverage for any_encode.go: every byte a struct plan or the
// pooled encoder produces must be what json.Marshal produces on this
// toolchain. The golden test freezes the format; these tests prove the
// fast paths reproduce it across the input space rather than at a few
// points.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"
)

// anyJSONValue logs v through Any on a JSON logger and returns the encoded
// value slot.
func anyJSONValue(t *testing.T, v any) string {
	t.Helper()
	var buf bytes.Buffer
	l := MustNew(Config{Format: FormatJSON, Writer: &buf, Level: LevelInfo})
	l.InfoEvent().Any("v", v).Msg("m")
	out := buf.String()
	const prefix = `{"level":"INFO","v":`
	const suffix = `,"message":"m"}` + "\n"
	if !strings.HasPrefix(out, prefix) || !strings.HasSuffix(out, suffix) {
		t.Fatalf("unexpected record %q", out)
	}
	return out[len(prefix) : len(out)-len(suffix)]
}

// wantMarshal is what the value slot must hold: json.Marshal's bytes, or
// the error form when it refuses the value.
func wantMarshal(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return string(appendJSONEscaped([]byte(`"!ERROR:`), err.Error())) + `"`
	}
	return string(b)
}

func planTableHas(t reflect.Type) (p *structPlan, seen bool) {
	m := structPlans.Load()
	if m == nil {
		return nil, false
	}
	p, seen = (*m)[t]
	return p, seen
}

type planBench struct {
	Rate string
	Low  int
	High float32
}

type planAllKinds struct {
	S    string
	B    bool
	I    int
	I8   int8
	I16  int16
	I32  int32
	I64  int64
	U    uint
	U8   uint8
	U16  uint16
	U32  uint32
	U64  uint64
	Uptr uintptr
	F32  float32
	F64  float64
}

type planRenamed struct {
	A string `json:"alpha"`
	B int    `json:"-"`
	C int    `json:"c-1_x"`
	D bool   `json:""`
	e int
	_ int
	F float64
}

type planLevel int

func (l planLevel) String() string { return "lvl" }

type planNamedKinds struct {
	L planLevel
	D time.Duration
	S planStr
}

type planStr string

// The benchmark type, every primitive kind, tag renames, unexported and
// blank fields and named primitives all get a plan, and the plan must agree
// with json.Marshal on the values that exercise it.
func TestStructPlanMatchesEncodingJSON(t *testing.T) {
	bench := planBench{"15", 16, 123.2}
	values := []any{
		&bench, bench, planBench{}, (*planBench)(nil), &planBench{Rate: "<&>\u2028\xff\"\\\n", Low: math.MinInt64, High: 1e-7},
		planAllKinds{S: "s", B: true, I: -1, I8: -128, I16: -32768, I32: math.MinInt32, I64: math.MinInt64, U: 1, U8: 255, U16: 65535, U32: math.MaxUint32, U64: math.MaxUint64, Uptr: 42, F32: 123.2, F64: 1e21},
		&planAllKinds{F32: 1e-7, F64: 0.000001},
		planRenamed{A: "a", B: 1, C: 2, D: true, e: 3, F: 0.5},
		&planRenamed{},
		planNamedKinds{L: 4, D: 1500 * time.Millisecond, S: "named"},
		struct{}{},
		&struct{ Only string }{"one"},
		struct {
			Größe   int
			Ünïcode string
		}{1, "ü"},
	}
	for _, v := range values {
		if got, want := anyJSONValue(t, v), wantMarshal(v); got != want {
			t.Errorf("%T: got %s want %s", v, got, want)
		}
		if p, seen := planTableHas(reflect.TypeOf(v)); !seen || p == nil {
			t.Errorf("%T: expected a plan, seen=%v plan=%v", v, seen, p != nil)
		}
	}
}

type planEmbedded struct {
	planBench
	X int
}

type planNested struct {
	Inner planBench
}

type planPtrField struct {
	P *int
}

type planIfaceField struct {
	I any
}

type planSliceField struct {
	L []int
}

type planOmitEmpty struct {
	A string `json:"a,omitempty"`
}

type planStringTag struct {
	A int `json:",string"`
}

type planQuotedTag struct {
	A int `json:"'a b'"`
}

type planDuplicate struct {
	A int
	B int `json:"A"`
}

type planFoldDuplicate struct {
	A int
	B int `json:"a"`
}

type planNumber struct {
	N json.Number
}

type planTime struct {
	T time.Time
}

type planValueMarshaler struct{ N int }

func (planValueMarshaler) MarshalJSON() ([]byte, error) { return []byte(`"vm"`), nil }

type planPtrMarshaler struct{ N int }

func (*planPtrMarshaler) MarshalJSON() ([]byte, error) { return []byte(`"pm"`), nil }

type planTextMarshaler struct{ N int }

func (planTextMarshaler) MarshalText() ([]byte, error) { return []byte("tm"), nil }

type planAppendText struct{ N int }

func (planAppendText) AppendText(b []byte) ([]byte, error) { return append(b, "at"...), nil }

type planMarshalerField struct {
	M planTextMarshaler
}

type planPtrMarshalerField struct {
	M planPtrMarshaler
}

// Types with any feature the plan does not reproduce are recorded as having
// no plan and still encode exactly as json.Marshal does, through the pooled
// encoder.
func TestStructPlanFallsBackForUnsupportedTypes(t *testing.T) {
	values := []any{
		planEmbedded{planBench{"r", 1, 2}, 3},
		planNested{planBench{"r", 1, 2}},
		planPtrField{},
		planIfaceField{I: 1},
		planSliceField{L: []int{1}},
		planOmitEmpty{}, planOmitEmpty{A: "x"},
		planStringTag{A: 5},
		planQuotedTag{A: 5},
		planDuplicate{1, 2},
		planFoldDuplicate{1, 2},
		planNumber{N: "1.5"},
		planTime{T: time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)},
		planValueMarshaler{1}, &planValueMarshaler{1},
		planPtrMarshaler{1}, &planPtrMarshaler{1},
		planTextMarshaler{1}, &planTextMarshaler{1},
		planAppendText{1},
		planMarshalerField{}, &planMarshalerField{},
		planPtrMarshalerField{}, &planPtrMarshalerField{},
	}
	for _, v := range values {
		if got, want := anyJSONValue(t, v), wantMarshal(v); got != want {
			t.Errorf("%T: got %s want %s", v, got, want)
		}
		if p, seen := planTableHas(reflect.TypeOf(v)); !seen || p != nil {
			t.Errorf("%T: expected no plan, seen=%v plan=%v", v, seen, p != nil)
		}
	}
	// non-struct types never enter the table
	for _, v := range []any{map[string]int{"a": 1}, []int{1}, 3, "s", new(int), [1]int{1}} {
		if got, want := anyJSONValue(t, v), wantMarshal(v); got != want {
			t.Errorf("%T: got %s want %s", v, got, want)
		}
		if _, seen := planTableHas(reflect.TypeOf(v)); seen {
			t.Errorf("%T: entered the plan table", v)
		}
	}
}

type planNaN struct {
	Rate string
	F    float64
	G    float32
}

// NaN and infinities are refused with encoding/json's own wording, in both
// formats, and the record around them is intact.
func TestStructPlanRefusesNonFiniteFloats(t *testing.T) {
	for _, v := range []any{
		planNaN{"r", math.NaN(), 1}, &planNaN{"r", 1, float32(math.Inf(1))}, planNaN{"r", math.Inf(-1), 0},
	} {
		if got, want := anyJSONValue(t, v), wantMarshal(v); got != want {
			t.Errorf("%+v: got %s want %s", v, got, want)
		}
		var buf bytes.Buffer
		l := MustNew(Config{Format: FormatConsole, Writer: &buf, Level: LevelInfo, DisableColor: true})
		l.InfoEvent().Any("v", v).Msg("m")
		_, err := json.Marshal(v)
		if want := "INFO v=!ERROR:" + err.Error() + " m\n"; buf.String() != want {
			t.Errorf("console %+v: got %q want %q", v, buf.String(), want)
		}
	}
	// finite values of the same type still take the plan
	if p, seen := planTableHas(reflect.TypeOf(planNaN{})); !seen || p == nil {
		t.Fatal("planNaN should have a plan")
	}
	v := planNaN{"r", 0.5, 0.25}
	if got, want := anyJSONValue(t, v), wantMarshal(v); got != want {
		t.Errorf("got %s want %s", got, want)
	}
}

func stringCorpus() []string {
	corpus := []string{"", "plain", "<a href=\"x\">&amp;</a>", "\u2028\u2029", "\ufffd", "\xff", "\xed\xa0\x80", "\xc0\x80", "\xe0\x80\x80", "\xf0\x80\x80\x80", "\xf4\x90\x80\x80", "\xf8\x88\x80\x80\x80", "\xe6\x97", "😀", "\U0010ffff", "\x7f", "\x00\x1f\x20"}
	for c := 0; c < 256; c++ {
		corpus = append(corpus, string([]byte{byte(c)}), "a"+string([]byte{byte(c)})+"b", string([]byte{byte(c), byte(c)}))
	}
	for r := rune(0); r < 0x300; r++ {
		corpus = append(corpus, string(r))
	}
	for r := rune(0x2000); r < 0x2100; r++ {
		corpus = append(corpus, "x"+string(r)+"y")
	}
	for _, r := range []rune{0xd7ff, 0xe000, 0xfffd, 0xfffe, 0xffff, 0x10000, 0x10ffff} {
		corpus = append(corpus, string(r))
	}
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 20000; i++ {
		b := make([]byte, rng.Intn(12))
		for j := range b {
			b[j] = byte(rng.Intn(256))
		}
		corpus = append(corpus, string(b))
	}
	return corpus
}

func TestAppendMarshalStringMatchesEncodingJSON(t *testing.T) {
	for _, s := range stringCorpus() {
		want, _ := json.Marshal(s)
		if got := appendMarshalString(nil, s); string(got) != string(want) {
			t.Fatalf("%q: got %s want %s", s, got, want)
		}
	}
}

func TestAppendMarshalFloatMatchesEncodingJSON(t *testing.T) {
	floats := []float64{0, math.Copysign(0, -1), 1, -1, 0.5, 0.1, 1.0 / 3, 123.2, 1e-7, 9.99e-7, 1e-6, 1.0000001e-6, 1e20, 999999999999999999999, 1e21, 1.5e-9, 1e-100, 1e100, 1e300, math.MaxFloat64, math.SmallestNonzeroFloat64, math.MaxFloat32, math.SmallestNonzeroFloat32, 5e-324, 2.2250738585072014e-308, 16777216, 16777217}
	rng := rand.New(rand.NewSource(2))
	for i := 0; i < 100000; i++ {
		f := math.Float64frombits(rng.Uint64())
		if math.IsNaN(f) || math.IsInf(f, 0) {
			continue
		}
		floats = append(floats, f)
	}
	for i := 0; i < 100000; i++ {
		f := math.Float32frombits(rng.Uint32())
		if math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) {
			continue
		}
		floats = append(floats, float64(f))
	}
	for _, f := range floats {
		want, _ := json.Marshal(f)
		if got := appendMarshalFloat(nil, f, 64); string(got) != string(want) {
			t.Fatalf("float64 %v: got %s want %s", f, got, want)
		}
		f32 := float32(f)
		if math.IsInf(float64(f32), 0) {
			continue
		}
		want, _ = json.Marshal(f32)
		if got := appendMarshalFloat(nil, float64(f32), 32); string(got) != string(want) {
			t.Fatalf("float32 %v: got %s want %s", f32, got, want)
		}
	}
}

// Whole records for random struct contents, by value and by pointer, to tie
// the field readers, keys and primitive encoders together.
func TestStructPlanRandomValuesMatchEncodingJSON(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	corpus := stringCorpus()
	for i := 0; i < 3000; i++ {
		v := planAllKinds{
			S: corpus[rng.Intn(len(corpus))], B: rng.Intn(2) == 0,
			I: int(rng.Uint64()), I8: int8(rng.Uint32()), I16: int16(rng.Uint32()), I32: int32(rng.Uint32()), I64: int64(rng.Uint64()),
			U: uint(rng.Uint64()), U8: uint8(rng.Uint32()), U16: uint16(rng.Uint32()), U32: rng.Uint32(), U64: rng.Uint64(), Uptr: uintptr(rng.Uint64()),
			F32: math.Float32frombits(rng.Uint32()), F64: math.Float64frombits(rng.Uint64()),
		}
		for _, x := range []any{v, &v} {
			if got, want := anyJSONValue(t, x), wantMarshal(x); got != want {
				t.Fatalf("%+v: got %s want %s", v, got, want)
			}
		}
	}
}

// Struct values through Any allocate nothing once their plan exists, in
// both the typed and the slog entry points; the pooled encoder also frees
// the fallback of json.Marshal's result allocation for a nested struct.
func TestAnyStructDoesNotAllocate(t *testing.T) {
	if raceEnabled {
		t.Skip("the race detector allocates for its own bookkeeping")
	}
	l := MustNew(Config{Format: FormatJSON, Writer: discardWriter{}, Level: LevelInfo})
	obj := planBench{"15", 16, 123.2}
	// Converting a struct to any boxes it, which is the caller's allocation,
	// not the encoder's; box once so the measurement covers only encoding.
	var boxed any = obj
	nested := planNested{obj}
	for name, fn := range map[string]func(){
		"pointer": func() {
			l.InfoEvent().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Any("object", &obj).Msg("m")
		},
		"value":  func() { l.InfoEvent().Any("object", boxed).Msg("m") },
		"nested": func() { l.InfoEvent().Any("object", &nested).Msg("m") },
	} {
		fn() // build the plan
		if got := testing.AllocsPerRun(1000, fn); got != 0 {
			t.Errorf("%s allocated %.2f times", name, got)
		}
	}
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }

// An encoder that refused a value is reused for the next one.
func TestPooledEncoderSurvivesRefusals(t *testing.T) {
	cyclic := []any{nil}
	cyclic[0] = cyclic
	for i := 0; i < 50; i++ {
		if got, want := anyJSONValue(t, cyclic), wantMarshal(cyclic); got != want {
			t.Fatalf("cycle: got %s want %s", got, want)
		}
		v := planNested{planBench{"r", i, 1}}
		if got, want := anyJSONValue(t, v), wantMarshal(v); got != want {
			t.Fatalf("after refusal: got %s want %s", got, want)
		}
	}
}

// The table stops growing at maxStructPlans and never blocks or corrupts
// under concurrent first sightings; types past the bound still encode
// correctly through the encoder.
func TestStructPlanTableIsBoundedAndConcurrent(t *testing.T) {
	// Filling the table is global; hand the previous table back afterwards
	// so later tests still see their types planned.
	savedTable, savedFull := structPlans.Load(), structPlansFull.Load()
	t.Cleanup(func() {
		structPlansMu.Lock()
		defer structPlansMu.Unlock()
		structPlans.Store(savedTable)
		structPlansFull.Store(savedFull)
	})
	types := make([]reflect.Type, 0, maxStructPlans+64)
	for i := 0; i < maxStructPlans+64; i++ {
		types = append(types, reflect.StructOf([]reflect.StructField{
			{Name: "N", Type: reflect.TypeFor[int]()},
			{Name: fmt.Sprintf("F%d", i), Type: reflect.TypeFor[string]()},
		}))
	}
	var wg sync.WaitGroup
	for g := 0; g < runtime.GOMAXPROCS(0)*2; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			var buf bytes.Buffer
			l := MustNew(Config{Format: FormatJSON, Writer: &buf, Level: LevelInfo})
			for i, ty := range types {
				v := reflect.New(ty)
				v.Elem().Field(0).SetInt(int64(i))
				v.Elem().Field(1).SetString("g")
				for _, x := range []any{v.Interface(), v.Elem().Interface()} {
					buf.Reset()
					l.InfoEvent().Any("v", x).Msg("m")
					want, _ := json.Marshal(x)
					if !bytes.Contains(buf.Bytes(), append([]byte(`"v":`), want...)) {
						t.Errorf("goroutine %d type %d: %s lacks %s", g, i, buf.Bytes(), want)
						return
					}
				}
			}
		}(g)
	}
	wg.Wait()
	if m := structPlans.Load(); m == nil || len(*m) > maxStructPlans {
		t.Fatalf("plan table holds %d entries, bound is %d", len(*m), maxStructPlans)
	}
	if !structPlansFull.Load() {
		t.Fatal("table should report full")
	}
}

// The plan's probe rejects a plan that does not reproduce json.Marshal. The
// probe string carries an ill-formed byte, so a toolchain whose encoding/json
// spells U+FFFD differently would refuse every plan with a string field
// rather than diverge.
func TestStructPlanProbeCoversEscaping(t *testing.T) {
	probe := "pr\"\\<&>\n\t\u2028\xffobe"
	if utf8.ValidString(probe) {
		t.Fatal("probe must contain ill-formed UTF-8")
	}
	want, _ := json.Marshal(probe)
	if got := appendMarshalString(nil, probe); string(got) != string(want) {
		t.Fatalf("got %s want %s", got, want)
	}
}
