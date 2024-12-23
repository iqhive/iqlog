package iqlog

func (l *logger) WithField(name string, value any) *logger {
	if !l.Enabled(l.ctx, LevelDebug) {
		return l
	}
	nl := l
	if nl.baseRecord.UsedFields < maxFields {
		i := nl.baseRecord.UsedFields
		nl.baseRecord.SetField(i, name, value)
		nl.baseRecord.UsedFields++
	}
	return nl
}

func (l *logger) WithFields(fields map[string]any) Logger {
	if !l.Enabled(l.ctx, LevelDebug) {
		return l
	}
	nl := l // single copy, done once
	i := nl.baseRecord.UsedFields
	for k, v := range fields {
		if i >= maxFields {
			break
		}
		nl.baseRecord.SetField(i, k, v)
		i++
	}
	nl.baseRecord.UsedFields = i
	return nl
}

func (l *logger) Int(name string, val int) *logger {
	if !l.Enabled(l.ctx, LevelDebug) {
		return l
	}
	nl := l
	if nl.baseRecord.UsedFields < maxFields {
		i := nl.baseRecord.UsedFields
		nl.baseRecord.SetField(i, name, val)
		nl.baseRecord.UsedFields++
	}
	return nl
}

func (l *logger) Str(name string, s string) *logger {
	if !l.Enabled(l.ctx, LevelDebug) {
		return l
	}
	nl := l
	if nl.baseRecord.UsedFields < maxFields {
		i := nl.baseRecord.UsedFields
		nl.baseRecord.SetField(i, name, s)
		nl.baseRecord.UsedFields++
	}
	return nl
}

// func (l *logger) WithStr(name string, s string) *LogItem {
// 	nl := l
// 	if nl.baseRecord.UsedFields < maxFields {
// 		i := nl.baseRecord.UsedFields
// 		nl.baseRecord.SetField(i, name, s)
// 		nl.baseRecord.UsedFields++
// 	}
// 	return nl
// }

func (l *logger) Float32(name string, f float32) *logger {
	if !l.Enabled(l.ctx, LevelDebug) {
		return l
	}
	nl := l
	if nl.baseRecord.UsedFields < maxFields {
		i := nl.baseRecord.UsedFields
		nl.baseRecord.SetField(i, name, f)
		nl.baseRecord.UsedFields++
	}
	return nl
}

func (l *logger) Float64(name string, f float64) *logger {
	if !l.Enabled(l.ctx, LevelDebug) {
		return l
	}
	nl := l
	if nl.baseRecord.UsedFields < maxFields {
		i := nl.baseRecord.UsedFields
		nl.baseRecord.SetField(i, name, f)
		nl.baseRecord.UsedFields++
	}
	return nl
}

func (l *logger) Bool(name string, b bool) *logger {
	if !l.Enabled(l.ctx, LevelDebug) {
		return l
	}
	nl := l
	if nl.baseRecord.UsedFields < maxFields {
		i := nl.baseRecord.UsedFields
		nl.baseRecord.SetField(i, name, b)
		nl.baseRecord.UsedFields++
	}
	return nl
}

func (l *logger) Any(name string, v any) *logger {
	if !l.Enabled(l.ctx, LevelDebug) {
		return l
	}
	nl := l
	if nl.baseRecord.UsedFields < maxFields {
		i := nl.baseRecord.UsedFields
		nl.baseRecord.SetField(i, name, v)
		nl.baseRecord.UsedFields++
	}
	return nl
}

func (l *logger) WithError(err error) Logger {
	if !l.Enabled(l.ctx, LevelDebug) || err == nil {
		return l
	}
	return l.WithField("error", err.Error())
}
