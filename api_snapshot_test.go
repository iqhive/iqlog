package iqlog

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strings"
	"testing"
)

func TestExportedAPISnapshot(t *testing.T) {
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, ".", func(info os.FileInfo) bool {
		return strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	var actual []string
	for _, file := range packages["iqlog"].Files {
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.FuncDecl:
				if ast.IsExported(decl.Name.Name) {
					prefix := "func "
					if decl.Recv != nil {
						receiver := decl.Recv.List[0].Type
						if pointer, ok := receiver.(*ast.StarExpr); ok {
							receiver = pointer.X
						}
						name, ok := receiver.(*ast.Ident)
						if !ok || !ast.IsExported(name.Name) {
							continue
						}
						prefix = "method "
					}
					actual = append(actual, prefix+decl.Name.Name)
				}
			case *ast.GenDecl:
				for _, spec := range decl.Specs {
					switch spec := spec.(type) {
					case *ast.TypeSpec:
						if ast.IsExported(spec.Name.Name) {
							actual = append(actual, "type "+spec.Name.Name)
						}
					case *ast.ValueSpec:
						for _, name := range spec.Names {
							if ast.IsExported(name.Name) {
								actual = append(actual, decl.Tok.String()+" "+name.Name)
							}
						}
					}
				}
			}
		}
	}
	sort.Strings(actual)
	canonical := []string{
		"type Config", "type Event", "type Format", "type JSONTimeMode", "type Level", "type Logger", "type OverflowPolicy", "type WriterMode",
		"var ErrClosed", "var ErrWriteDropped", "const FormatConsole", "const FormatJSON", "const JSONTimeUTC", "const JSONTimeDisabled", "const JSONTimeCustom", "const OverflowBlock", "const OverflowDrop", "const OverflowSync", "const WriterAsync", "const WriterRing", "const WriterSync",
		"const LevelUnknown", "const LevelTrace", "const LevelDebug", "const LevelInfo", "const LevelWarn", "const LevelError", "const LevelPanic", "const LevelFatal",
		"func New", "func MustNew", "func Default", "func SetDefault", "func Enabled", "func Dropped", "func ParseLevel", "func LastWriteError",
		"func Trace", "func Tracef", "func Traceln", "func Debug", "func Debugf", "func Debugln", "func Info", "func Infof", "func Infoln", "func Warn", "func Warnf", "func Warnln", "func Error", "func Errorf", "func Errorln", "func Panic", "func Panicf", "func Panicln", "func Fatal", "func Fatalf", "func Fatalln", "func Print", "func Printf", "func Println", "func Log", "func Logf", "func Logln", "func LogContext", "func LogContextf", "func WithContext", "func WithFields", "func WithError",
		"func TraceEvent", "func DebugEvent", "func InfoEvent", "func WarnEvent", "func ErrorEvent", "func PanicEvent", "func FatalEvent",
		"method Config", "method SetConfig", "method Level", "method Enabled", "method Dropped", "method LastWriteError", "method Flush", "method Close", "method WithContext", "method WithFields", "method WithError", "method Event", "method EventAt", "method TraceEvent", "method DebugEvent", "method InfoEvent", "method WarnEvent", "method ErrorEvent", "method PanicEvent", "method FatalEvent",
		"method Trace", "method Tracef", "method Traceln", "method Debug", "method Debugf", "method Debugln", "method Info", "method Infof", "method Infoln", "method Warn", "method Warnf", "method Warnln", "method Error", "method Errorf", "method Errorln", "method Panic", "method Panicf", "method Panicln", "method Fatal", "method Fatalf", "method Fatalln", "method Print", "method Printf", "method Println", "method Log", "method Logf", "method Logln", "method LogContext", "method LogContextf",
		"method Any", "method Bool", "method BuildError", "method Bytes", "method Discard", "method Duration", "method Err", "method Float32", "method Float64", "method Int", "method Int64", "method RawJSON", "method Str", "method Stringer", "method Time", "method Uint", "method Uint64", "method Msg", "method Msgs", "method Msgf", "method String", "method MarshalText", "method UnmarshalText",
	}
	legacy := []string{
		"const LevelPrint", "func NewIQLogger", "func NewGlobalIQLogger", "func Init", "func Warning", "func Warningf", "func Warningln", "func TraceWith", "func DebugWith", "func InfoWith", "func WarnWith", "func ErrorWith", "func PanicWith", "func FatalWith", "func SetWriter", "func GetWriter", "func SetLevel", "func SetDebugMode", "func SetCallerDepth", "func SetUseColour", "func SetUseColor", "func SetJSONMode", "func SetNewLine", "func SetApplicationName", "func SetSyslogHost", "func Flush",
		"method Warning", "method Warningf", "method Warningln", "method TraceWith", "method DebugWith", "method InfoWith", "method WarnWith", "method ErrorWith", "method PanicWith", "method FatalWith", "method SetWriter", "method SetAsyncWriter", "method SetRingbufferWriter", "method SetRingBufferWriter", "method GetWriter", "method SetLevel", "method SetDebugMode", "method SetCallerDepth", "method SetUseColour", "method SetUseColor", "method SetJSONMode", "method SetNewLine", "method SetApplicationName", "method SetSyslogHost", "method LogWithFields", "method LogFWithFields",
	}
	want := append(canonical, legacy...)
	sort.Strings(want)
	if strings.Join(actual, "\n") != strings.Join(want, "\n") {
		t.Fatalf("exported API changed\nactual:\n%s\n\nwant:\n%s", strings.Join(actual, "\n"), strings.Join(want, "\n"))
	}
}
