package iqlog

// Golden coverage for Event.Any and the generic value encoder it falls back
// to. The expected strings were captured from the encoder as it stood before
// the per-type struct plans and the pooled json.Encoder were introduced, and
// they are frozen here as literals: the record formats are a public contract,
// so any change to these bytes is a breaking change, not a refactor.
//
// The table covers the comparative benchmark payload (bench/BenchmarkIQLogAny)
// and its neighbours: the same struct by value, nil pointers, tags, unexported
// and blank fields, nested and container fields, every string-escaping class
// encoding/json distinguishes, float formatting at its 'e'/'f' thresholds,
// integer extremes, named primitive types, the values encoding/json refuses
// (NaN, Inf, cycles) and the custom marshalers it defers to, including one
// that returns ill-formed UTF-8.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"strings"
	"testing"
	"time"
)

const goldenMsg = "The quick brown fox jumps over the lazy dog"

// goldenBenchObj mirrors bench/benchmark_test.go's obj exactly, including its
// anonymous type.
var goldenBenchObj = struct {
	Rate string
	Low  int
	High float32
}{"15", 16, 123.2}

type goldenNamed struct {
	Rate string
	Low  int
	High float32
}

type goldenTagged struct {
	Rate    string `json:"rate"`
	Skipped int    `json:"-"`
	Dash    int    `json:"-,"`
	Empty   string `json:",omitempty"`
	Zero    int    `json:"zero,omitempty"`
	Quoted  int    `json:"quoted,string"`
	NoTag   string `json:""`
	hidden  int
	_       int
	Last    bool
}

type goldenNested struct {
	Inner goldenNamed
	Ptr   *goldenNamed
	Nil   *goldenNamed
	List  []int
	Map   map[string]int
	Iface any
	Arr   [2]bool
	Bytes []byte
}

type goldenEmbedded struct {
	goldenNamed
	Extra string
}

type goldenLevel int

func (l goldenLevel) String() string { return fmt.Sprintf("level-%d", int(l)) }

type goldenPrimitives struct {
	I8   int8
	I16  int16
	I32  int32
	I64  int64
	I    int
	U8   uint8
	U16  uint16
	U32  uint32
	U64  uint64
	U    uint
	Uptr uintptr
	B    bool
	Lvl  goldenLevel
	Dur  time.Duration
}

type goldenFloats struct {
	Zero    float64
	NegZero float64
	Tiny    float64
	Edge    float64
	Big     float64
	BigEdge float64
	Half    float64
	F32     float32
	F32Tiny float32
	F32Max  float32
	F32Big  float32
	Exp     float64
}

type goldenStrings struct {
	HTML    string
	Quotes  string
	Control string
	DEL     string
	Unicode string
	LineSep string
	Bad     string
	Empty   string
}

type goldenNaN struct {
	Rate string
	F    float64
}

type goldenInf struct {
	F32 float32
}

type goldenValueMarshaler struct{ N int }

func (m goldenValueMarshaler) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"custom":%d}`, m.N)), nil
}

type goldenPtrMarshaler struct{ N int }

func (m *goldenPtrMarshaler) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"ptr":%d}`, m.N)), nil
}

type goldenTextMarshaler struct{ N int }

func (m goldenTextMarshaler) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("text-%d", m.N)), nil
}

type goldenBadUTF8Marshaler struct{}

func (goldenBadUTF8Marshaler) MarshalJSON() ([]byte, error) { return []byte("\"\xd9\xff\""), nil }

type goldenFailingMarshaler struct{}

func (goldenFailingMarshaler) MarshalJSON() ([]byte, error) { return nil, errors.New("refused") }

type goldenWithNumber struct {
	N json.Number
	T time.Time
}

type goldenErr struct{ Code int }

func (e *goldenErr) Error() string { return fmt.Sprintf("code %d", e.Code) }

type goldenCase struct {
	name        string
	build       func(e *Event) *Event
	wantJSON    string
	wantConsole string
}

func goldenCases() []goldenCase {
	cyclicSlice := []any{nil}
	cyclicSlice[0] = cyclicSlice
	cyclicMap := map[string]any{}
	cyclicMap["self"] = cyclicMap
	named := goldenNamed{"15", 16, 123.2}
	var nilNamed *goldenNamed
	var nilErr *goldenErr
	return []goldenCase{
		{
			name: "benchmark payload",
			build: func(e *Event) *Event {
				return e.Str("rate", "15").Int("low", 16).Float32("high", 123.2).Any("object", &goldenBenchObj)
			},
			wantJSON:    "{\"level\":\"INFO\",\"rate\":\"15\",\"low\":16,\"high\":123.2,\"object\":{\"Rate\":\"15\",\"Low\":16,\"High\":123.2},\"message\":\"The quick brown fox jumps over the lazy dog\"}\n",
			wantConsole: "INFO rate=15 low=16 high=123.2 object=&{15 16 123.2} The quick brown fox jumps over the lazy dog\n",
		},
		{
			name:        "benchmark object by value",
			build:       func(e *Event) *Event { return e.Any("object", goldenBenchObj) },
			wantJSON:    "{\"level\":\"INFO\",\"object\":{\"Rate\":\"15\",\"Low\":16,\"High\":123.2},\"message\":\"The quick brown fox jumps over the lazy dog\"}\n",
			wantConsole: "INFO object={15 16 123.2} The quick brown fox jumps over the lazy dog\n",
		},
		{
			name:        "named struct by value and pointer",
			build:       func(e *Event) *Event { return e.Any("v", named).Any("p", &named) },
			wantJSON:    "{\"level\":\"INFO\",\"v\":{\"Rate\":\"15\",\"Low\":16,\"High\":123.2},\"p\":{\"Rate\":\"15\",\"Low\":16,\"High\":123.2},\"message\":\"The quick brown fox jumps over the lazy dog\"}\n",
			wantConsole: "INFO v={15 16 123.2} p=&{15 16 123.2} The quick brown fox jumps over the lazy dog\n",
		},
		{
			name:        "nil struct pointer",
			build:       func(e *Event) *Event { return e.Any("p", nilNamed) },
			wantJSON:    "{\"level\":\"INFO\",\"p\":null,\"message\":\"The quick brown fox jumps over the lazy dog\"}\n",
			wantConsole: "INFO p=<nil> The quick brown fox jumps over the lazy dog\n",
		},
		{
			name:        "zero struct",
			build:       func(e *Event) *Event { return e.Any("z", goldenNamed{}).Any("e", struct{}{}) },
			wantJSON:    "{\"level\":\"INFO\",\"z\":{\"Rate\":\"\",\"Low\":0,\"High\":0},\"e\":{},\"message\":\"The quick brown fox jumps over the lazy dog\"}\n",
			wantConsole: "INFO z={ 0 0} e={} The quick brown fox jumps over the lazy dog\n",
		},
		{
			name: "tags",
			build: func(e *Event) *Event {
				return e.Any("t", goldenTagged{Rate: "r", Skipped: 1, Dash: 2, Empty: "", Zero: 0, Quoted: 3, NoTag: "n", hidden: 4, Last: true}).
					Any("u", &goldenTagged{Empty: "x", Zero: 5})
			},
			wantJSON:    "{\"level\":\"INFO\",\"t\":{\"rate\":\"r\",\"-\":2,\"quoted\":\"3\",\"NoTag\":\"n\",\"Last\":true},\"u\":{\"rate\":\"\",\"-\":0,\"Empty\":\"x\",\"zero\":5,\"quoted\":\"0\",\"NoTag\":\"\",\"Last\":false},\"message\":\"The quick brown fox jumps over the lazy dog\"}\n",
			wantConsole: "INFO t={r 1 2  0 3 n 4 0 true} u=&{ 0 0 x 5 0  0 0 false} The quick brown fox jumps over the lazy dog\n",
		},
		{
			name: "nested and containers",
			build: func(e *Event) *Event {
				return e.Any("n", goldenNested{Inner: named, Ptr: &named, List: []int{1, 2}, Map: map[string]int{"b": 2, "a": 1}, Iface: "s", Arr: [2]bool{true, false}, Bytes: []byte("hi")})
			},
			wantJSON:    "{\"level\":\"INFO\",\"n\":{\"Inner\":{\"Rate\":\"15\",\"Low\":16,\"High\":123.2},\"Ptr\":{\"Rate\":\"15\",\"Low\":16,\"High\":123.2},\"Nil\":null,\"List\":[1,2],\"Map\":{\"a\":1,\"b\":2},\"Iface\":\"s\",\"Arr\":[true,false],\"Bytes\":\"aGk=\"},\"message\":\"The quick brown fox jumps over the lazy dog\"}\n",
			wantConsole: "INFO n={{15 16 123.2} 0xPTR <nil> [1 2] map[a:1 b:2] s [true false] [104 105]} The quick brown fox jumps over the lazy dog\n",
		},
		{
			name:        "embedded struct",
			build:       func(e *Event) *Event { return e.Any("e", goldenEmbedded{goldenNamed: named, Extra: "x"}) },
			wantJSON:    "{\"level\":\"INFO\",\"e\":{\"Rate\":\"15\",\"Low\":16,\"High\":123.2,\"Extra\":\"x\"},\"message\":\"The quick brown fox jumps over the lazy dog\"}\n",
			wantConsole: "INFO e={{15 16 123.2} x} The quick brown fox jumps over the lazy dog\n",
		},
		{
			name: "integer kinds and named primitives",
			build: func(e *Event) *Event {
				return e.Any("p", goldenPrimitives{I8: -128, I16: -32768, I32: math.MinInt32, I64: math.MinInt64, I: -1, U8: 255, U16: 65535, U32: math.MaxUint32, U64: math.MaxUint64, U: 7, Uptr: 9, B: true, Lvl: 3, Dur: 1500 * time.Millisecond})
			},
			wantJSON:    "{\"level\":\"INFO\",\"p\":{\"I8\":-128,\"I16\":-32768,\"I32\":-2147483648,\"I64\":-9223372036854775808,\"I\":-1,\"U8\":255,\"U16\":65535,\"U32\":4294967295,\"U64\":18446744073709551615,\"U\":7,\"Uptr\":9,\"B\":true,\"Lvl\":3,\"Dur\":1500000000},\"message\":\"The quick brown fox jumps over the lazy dog\"}\n",
			wantConsole: "INFO p={-128 -32768 -2147483648 -9223372036854775808 -1 255 65535 4294967295 18446744073709551615 7 9 true level-3 1.5s} The quick brown fox jumps over the lazy dog\n",
		},
		{
			name: "float formatting thresholds",
			build: func(e *Event) *Event {
				return e.Any("f", goldenFloats{Zero: 0, NegZero: math.Copysign(0, -1), Tiny: 1e-7, Edge: 1e-6, Big: 1e21, BigEdge: 999999999999999999999, Half: 0.5, F32: 123.2, F32Tiny: 1e-7, F32Max: math.MaxFloat32, F32Big: 1e21, Exp: 1.5e-9})
			},
			wantJSON:    "{\"level\":\"INFO\",\"f\":{\"Zero\":0,\"NegZero\":-0,\"Tiny\":1e-7,\"Edge\":0.000001,\"Big\":1e+21,\"BigEdge\":1e+21,\"Half\":0.5,\"F32\":123.2,\"F32Tiny\":1e-7,\"F32Max\":3.4028235e+38,\"F32Big\":1e+21,\"Exp\":1.5e-9},\"message\":\"The quick brown fox jumps over the lazy dog\"}\n",
			wantConsole: "INFO f={0 -0 1e-07 1e-06 1e+21 1e+21 0.5 123.2 1e-07 3.4028235e+38 1e+21 1.5e-09} The quick brown fox jumps over the lazy dog\n",
		},
		{
			name: "string escaping classes",
			build: func(e *Event) *Event {
				return e.Any("s", goldenStrings{HTML: "<a href=\"x\">&amp;</a>", Quotes: "q\"b\\s", Control: "n\nr\rt\tb\bf\fz\x00x\x1f", DEL: "d\x7fe", Unicode: "é日本😀", LineSep: "l\u2028s\u2029", Bad: "b\xffa\xd9d\xed\xa0\x80s", Empty: ""})
			},
			wantJSON:    "{\"level\":\"INFO\",\"s\":{\"HTML\":\"\\u003ca href=\\\"x\\\"\\u003e\\u0026amp;\\u003c/a\\u003e\",\"Quotes\":\"q\\\"b\\\\s\",\"Control\":\"n\\nr\\rt\\tb\\bf\\fz\\u0000x\\u001f\",\"DEL\":\"d\x7fe\",\"Unicode\":\"é日本😀\",\"LineSep\":\"l\\u2028s\\u2029\",\"Bad\":\"b�a�d���s\",\"Empty\":\"\"},\"message\":\"The quick brown fox jumps over the lazy dog\"}\n",
			wantConsole: "INFO s={<a href=\"x\">&amp;</a> q\"b\\s n\\nr\\rt\tb\\x08f\\x0cz\\x00x\\x1f d\\x7fe é日本😀 l\u2028s\u2029 b\xffa\xd9d\xed\xa0\\x80s } The quick brown fox jumps over the lazy dog\n",
		},
		{
			name:        "NaN is refused in json and console",
			build:       func(e *Event) *Event { return e.Any("n", goldenNaN{Rate: "r", F: math.NaN()}) },
			wantJSON:    "{\"level\":\"INFO\",\"n\":\"!ERROR:json: unsupported value: NaN\",\"message\":\"The quick brown fox jumps over the lazy dog\"}\n",
			wantConsole: "INFO n=!ERROR:json: unsupported value: NaN The quick brown fox jumps over the lazy dog\n",
		},
		{
			name: "infinities are refused",
			build: func(e *Event) *Event {
				return e.Any("p", &goldenInf{float32(math.Inf(1))}).Any("m", goldenInf{float32(math.Inf(-1))})
			},
			wantJSON:    "{\"level\":\"INFO\",\"p\":\"!ERROR:json: unsupported value: +Inf\",\"m\":\"!ERROR:json: unsupported value: -Inf\",\"message\":\"The quick brown fox jumps over the lazy dog\"}\n",
			wantConsole: "INFO p=!ERROR:json: unsupported value: +Inf m=!ERROR:json: unsupported value: -Inf The quick brown fox jumps over the lazy dog\n",
		},
		{
			name:        "cycles are refused before fmt sees them",
			build:       func(e *Event) *Event { return e.Any("s", cyclicSlice).Any("m", cyclicMap) },
			wantJSON:    "{\"level\":\"INFO\",\"s\":\"!ERROR:json: unsupported value: encountered a cycle via []interface {}\",\"m\":\"!ERROR:json: unsupported value: encountered a cycle via map[string]interface {}\",\"message\":\"The quick brown fox jumps over the lazy dog\"}\n",
			wantConsole: "INFO s=!ERROR:json: unsupported value: encountered a cycle via []interface {} m=!ERROR:json: unsupported value: encountered a cycle via map[string]interface {} The quick brown fox jumps over the lazy dog\n",
		},
		{
			name: "custom marshalers",
			build: func(e *Event) *Event {
				return e.Any("v", goldenValueMarshaler{1}).Any("pv", goldenPtrMarshaler{2}).Any("pp", &goldenPtrMarshaler{3}).Any("t", goldenTextMarshaler{4})
			},
			wantJSON:    "{\"level\":\"INFO\",\"v\":{\"custom\":1},\"pv\":{\"N\":2},\"pp\":{\"ptr\":3},\"t\":\"text-4\",\"message\":\"The quick brown fox jumps over the lazy dog\"}\n",
			wantConsole: "INFO v={1} pv={2} pp=&{3} t={4} The quick brown fox jumps over the lazy dog\n",
		},
		{
			name:        "marshaler returning ill-formed UTF-8 falls back to escaped text",
			build:       func(e *Event) *Event { return e.Any("b", goldenBadUTF8Marshaler{}) },
			wantJSON:    "{\"level\":\"INFO\",\"b\":\"{}\",\"message\":\"The quick brown fox jumps over the lazy dog\"}\n",
			wantConsole: "INFO b={} The quick brown fox jumps over the lazy dog\n",
		},
		{
			name:        "marshaler error",
			build:       func(e *Event) *Event { return e.Any("f", goldenFailingMarshaler{}) },
			wantJSON:    "{\"level\":\"INFO\",\"f\":\"!ERROR:json: error calling MarshalJSON for type *iqlog.goldenFailingMarshaler: refused\",\"message\":\"The quick brown fox jumps over the lazy dog\"}\n",
			wantConsole: "INFO f=!ERROR:json: error calling MarshalJSON for type *iqlog.goldenFailingMarshaler: refused The quick brown fox jumps over the lazy dog\n",
		},
		{
			name: "json.Number and time.Time fields",
			build: func(e *Event) *Event {
				return e.Any("n", goldenWithNumber{N: "12.5", T: time.Date(2024, 2, 3, 4, 5, 6, 7, time.UTC)})
			},
			wantJSON:    "{\"level\":\"INFO\",\"n\":{\"N\":12.5,\"T\":\"2024-02-03T04:05:06.000000007Z\"},\"message\":\"The quick brown fox jumps over the lazy dog\"}\n",
			wantConsole: "INFO n={12.5 2024-02-03 04:05:06.000000007 +0000 UTC} The quick brown fox jumps over the lazy dog\n",
		},
		{
			name: "containers and scalars through Any",
			build: func(e *Event) *Event {
				return e.Any("m", map[string]int{"z": 1, "a": 2}).Any("l", []string{"<", "&"}).Any("nil", nil).Any("i", 3).Any("f", 2.5).Any("b", true).Any("s", "x")
			},
			wantJSON:    "{\"level\":\"INFO\",\"m\":{\"a\":2,\"z\":1},\"l\":[\"\\u003c\",\"\\u0026\"],\"nil\":null,\"i\":3,\"f\":2.5,\"b\":true,\"s\":\"x\",\"message\":\"The quick brown fox jumps over the lazy dog\"}\n",
			wantConsole: "INFO m=map[a:2 z:1] l=[< &] nil=null i=3 f=2.5 b=true s=x The quick brown fox jumps over the lazy dog\n",
		},
		{
			name:        "errors including a typed nil pointer",
			build:       func(e *Event) *Event { return e.Any("e", &goldenErr{7}).Any("n", nilErr) },
			wantJSON:    "{\"level\":\"INFO\",\"e\":\"code 7\",\"n\":null,\"message\":\"The quick brown fox jumps over the lazy dog\"}\n",
			wantConsole: "INFO e=code 7 n=<nil> The quick brown fox jumps over the lazy dog\n",
		},
	}
}

func goldenLoggers(t *testing.T) (jsonLog *Logger, jsonBuf *bytes.Buffer, consoleLog *Logger, consoleBuf *bytes.Buffer) {
	t.Helper()
	jsonBuf, consoleBuf = &bytes.Buffer{}, &bytes.Buffer{}
	fixed := func() time.Time { return time.Date(2024, 1, 2, 3, 4, 5, 6000, time.UTC) }
	jsonLog = MustNew(Config{Format: FormatJSON, Writer: jsonBuf, Level: LevelInfo, JSONTimeMode: JSONTimeDisabled, Now: fixed})
	consoleLog = MustNew(Config{Format: FormatConsole, Writer: consoleBuf, Level: LevelInfo, DisableColor: true, Now: fixed})
	return
}

// normalizeGoldenPointer masks the one part of a console record that cannot
// be frozen: fmt prints a pointer field's address.
func normalizeGoldenPointer(s string) string {
	if i := strings.Index(s, " 0x"); i >= 0 {
		j := i + 3
		for j < len(s) && s[j] != ' ' {
			j++
		}
		return s[:i] + " 0xPTR" + s[j:]
	}
	return s
}

func TestAnyGolden(t *testing.T) {
	jsonLog, jsonBuf, consoleLog, consoleBuf := goldenLoggers(t)
	print := os.Getenv("IQLOG_GOLDEN_PRINT") != ""
	for _, tc := range goldenCases() {
		jsonBuf.Reset()
		consoleBuf.Reset()
		tc.build(jsonLog.InfoEvent()).Msg(goldenMsg)
		tc.build(consoleLog.InfoEvent()).Msg(goldenMsg)
		gotJSON := jsonBuf.String()
		gotConsole := normalizeGoldenPointer(consoleBuf.String())
		if print {
			fmt.Printf("\t\t{\n\t\t\tname: %q,\n\t\t\twantJSON:    %q,\n\t\t\twantConsole: %q,\n\t\t},\n", tc.name, gotJSON, gotConsole)
			continue
		}
		if gotJSON != tc.wantJSON {
			t.Errorf("%s: json\n got %q\nwant %q", tc.name, gotJSON, tc.wantJSON)
		}
		if gotConsole != tc.wantConsole {
			t.Errorf("%s: console\n got %q\nwant %q", tc.name, gotConsole, tc.wantConsole)
		}
	}
}

// The slog handler and the map-based legacy entry point share the generic
// encoder, so the benchmark object must render identically through them.
func TestAnyGoldenViaSlogAndFields(t *testing.T) {
	jsonLog, jsonBuf, consoleLog, consoleBuf := goldenLoggers(t)
	print := os.Getenv("IQLOG_GOLDEN_PRINT") != ""
	named := goldenNamed{"15", 16, 123.2}
	for _, tc := range []struct {
		name        string
		run         func(l *Logger)
		wantJSON    string
		wantConsole string
	}{
		{
			name: "slog",
			run: func(l *Logger) {
				slog.New(l.SlogHandler()).Info(goldenMsg, "object", &goldenBenchObj, "v", named, "n", goldenNaN{"r", math.NaN()})
			},
			wantJSON:    "{\"level\":\"INFO\",\"object\":{\"Rate\":\"15\",\"Low\":16,\"High\":123.2},\"v\":{\"Rate\":\"15\",\"Low\":16,\"High\":123.2},\"n\":\"!ERROR:json: unsupported value: NaN\",\"message\":\"The quick brown fox jumps over the lazy dog\"}\n",
			wantConsole: "INFO object=&{15 16 123.2} v={15 16 123.2} n=!ERROR:json: unsupported value: NaN The quick brown fox jumps over the lazy dog\n",
		},
		{
			name: "LogWithFields",
			run: func(l *Logger) {
				l.LogWithFields(context.Background(), LevelInfo, map[string]any{"object": &goldenBenchObj, "v": named}, goldenMsg)
			},
			wantJSON:    "{\"level\":\"INFO\",\"object\":{\"Rate\":\"15\",\"Low\":16,\"High\":123.2},\"v\":{\"Rate\":\"15\",\"Low\":16,\"High\":123.2},\"message\":\"The quick brown fox jumps over the lazy dog\"}\n",
			wantConsole: "INFO object=&{15 16 123.2} v={15 16 123.2} The quick brown fox jumps over the lazy dog\n",
		},
	} {
		jsonBuf.Reset()
		consoleBuf.Reset()
		tc.run(jsonLog)
		tc.run(consoleLog)
		gotJSON, gotConsole := jsonBuf.String(), consoleBuf.String()
		if print {
			fmt.Printf("\t\t{\n\t\t\tname: %q,\n\t\t\twantJSON:    %q,\n\t\t\twantConsole: %q,\n\t\t},\n", tc.name, gotJSON, gotConsole)
			continue
		}
		if gotJSON != tc.wantJSON {
			t.Errorf("%s: json\n got %q\nwant %q", tc.name, gotJSON, tc.wantJSON)
		}
		if gotConsole != tc.wantConsole {
			t.Errorf("%s: console\n got %q\nwant %q", tc.name, gotConsole, tc.wantConsole)
		}
	}
}
