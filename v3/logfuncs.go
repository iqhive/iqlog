package iqlog

// Enabled reports whether the handler handles records at the given level.
// The handler ignores records whose level is lower.
// It is called early, before any arguments are processed,
// to save effort if the log event should be discarded.
// If called from a Logger method, the first argument is the context
// passed to that method, or context.Background() if nil was passed
// or the method does not take a context.
// The context is passed so Enabled can use its values
// to make a decision.
func (l logger) Enabled(level Level) bool {
	// TODO: use context to check if the logger is enabled
	return level >= l.Level
}

// func (l *preallocLine) Msg(format string, args ...interface{}) {
// 	if l.Level > LevelTrace {
// 		return
// 	}
// 	if len(args) == 0 {
// 		l.HandleMsg(LevelTrace, format)
// 	} else {
// 		l.HandleMsgf(LevelTrace, format, args...)
// 	}
// }

// func (l *preallocLine) Msgf(msg string, args ...interface{}) {
// 	if l.Level > LevelTrace {
// 		return
// 	}
// 	if len(args) == 0 {
// 		l.HandleMsg(LevelTrace, msg)
// 	} else {
// 		l.HandleMsg(LevelTrace, msg, args...)
// 	}
// }
