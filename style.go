package iqlog

import "log/slog"

const (
	grey   = 0
	red    = 31
	green  = 32
	yellow = 33
	blue   = 36
	white  = 37
)

// // Handle handles the Record.
// // It will only be called when Enabled returns true.
// // The Context argument is as for Enabled.
// // It is present solely to provide Handlers access to the context's values.
// // Canceling the context should not affect record processing.
// // (Among other things, log messages may be necessary to debug a
// // cancellation-related problem.)
// //
// // Handle methods that produce output should observe the following rules:
// //   - If r.Time is the zero time, ignore the time.
// //   - If r.PC is zero, ignore it.
// //   - Attr's values should be resolved.
// //   - If an Attr's key and value are both the zero value, ignore the Attr.
// //     This can be tested with attr.Equal(Attr{}).
// //   - If a group's key is empty, inline the group's Attrs.
// //   - If a group has no Attrs (even if it has a non-empty key),
// //     ignore it.
// func (l logger) Handle(ctx context.Context, entry slog.Record) error {
// 	var b = new(bytes.Buffer)
// 	var (
// 		levelText   = "????"
// 		traceColour = green
// 	)
// 	var levelColor int
// 	switch Level(int(entry.Level)) {
// 	case LevelInfo:
// 		levelColor = blue
// 		levelText = "INFO"
// 	case LevelDebug, LevelTrace:
// 		levelColor = white
// 		traceColour = grey
// 		levelText = "DBUG"
// 	case LevelWarn:
// 		levelColor = yellow
// 		levelText = "WARN"
// 	case LevelError:
// 		levelColor = red
// 		levelText = "ERRR"
// 	case LevelFatal:
// 		levelColor = red
// 		levelText = "FATL"
// 	case LevelPanic:
// 		levelColor = red
// 		levelText = "PANC"
// 	default:
// 		levelColor = blue
// 	}
// 	if l.IncludeTimePrefix {
// 		if l.TimePrefixFormat == "" {
// 			fmt.Fprintf(b, "%s ", entry.Time.Format(time.StampMicro))
// 		} else {
// 			fmt.Fprintf(b, "%s ", entry.Time.Format(l.TimePrefixFormat))
// 		}
// 	}
// 	if l.useColour {
// 		fmt.Fprintf(b, "\x1b[%dm%s\x1b[0m ", levelColor, levelText)
// 	} else {
// 		fmt.Fprintf(b, "%s ", levelText)
// 	}
// 	originText := ""
// 	if entry.PC != 0 {
// 		var callers [6]uintptr
// 		runtime.Callers(3, callers[:])
// 		var source *runtime.Func
// 		var file string
// 		var line int
// 		for i := range callers {
// 			source = runtime.FuncForPC(callers[i])
// 			name := source.Name()
// 			file, line = source.FileLine(callers[i])
// 			if strings.HasSuffix(file, "v3/iqlog.go") || strings.HasSuffix(file, "v3/apierror.go") || strings.Contains(name, "Log") || strings.Contains(name, "APIError") {
// 				continue
// 			}
// 			break
// 		}
// 		name := path.Base(KeepNumDirs(source.Name(), NumPathsToLog))
// 		file = KeepNumDirs(file, NumPathsToLog)
// 		originText = fmt.Sprintf("[%v %v:%v]", name, file, line)
// 	} else {
// 		originText = fmt.Sprintf("%v", l.applicationName)
// 	}
// 	if originText != "" {
// 		if l.useColour {
// 			fmt.Fprintf(b, "\x1b[%dm%s\x1b[0m ", traceColour, originText)
// 		} else {
// 			fmt.Fprintf(b, "%s ", originText)
// 		}
// 	}
// 	fmt.Fprintf(b, "%s ", entry.Message)
// 	b.WriteByte('\n')
// 	_, err := os.Stderr.Write(b.Bytes())
// 	return err
// }

// WithAttrs returns a new logger whose attributes consist of
// both the receiver's attributes and the arguments.
func (l *logger) WithAttrs(attrs []slog.Attr) *logger {
	return l
}

// // WithGroup returns a new Handler with the given group appended to
// // the receiver's existing groups.
// // The keys of all subsequent attributes, whether added by With or in a
// // Record, should be qualified by the sequence of group names.
// //
// // How this qualification happens is up to the Handler, so long as
// // this Handler's attribute keys differ from those of another Handler
// // with a different sequence of group names.
// //
// // A Handler should treat WithGroup as starting a Group of Attrs that ends
// // at the end of the log event. That is,
// //
// //	logger.WithGroup("s").LogAttrs(ctx, level, msg, slog.Int("a", 1), slog.Int("b", 2))
// //
// // should behave like
// //
// //	logger.LogAttrs(ctx, level, msg, slog.Group("s", slog.Int("a", 1), slog.Int("b", 2)))
// //
// // If the name is empty, WithGroup returns the receiver.
// func (l logger) WithGroup(name string) slog.Handler {
// 	l.slog = &slogEmu{Logger: l.slog.Logger.WithGroup(name)}
// 	return l.slog
// }
