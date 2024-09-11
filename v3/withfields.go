package iqlog

func (l *logger) WithField(name string, value any) *logger {
	nl := l
	if nl.baseRecord.UsedFields < maxFields {
		i := nl.baseRecord.UsedFields
		nl.baseRecord.SetField(i, name, value)
		nl.baseRecord.UsedFields++
	}
	return nl
}

func (l *logger) WithFields(fields map[string]any) Logger {
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
	nl := l
	if nl.baseRecord.UsedFields < maxFields {
		i := nl.baseRecord.UsedFields
		nl.baseRecord.SetField(i, name, val)
		nl.baseRecord.UsedFields++
	}
	return nl
}

func (l *logger) Str(name string, s string) *logger {
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
	nl := l
	if nl.baseRecord.UsedFields < maxFields {
		i := nl.baseRecord.UsedFields
		nl.baseRecord.SetField(i, name, f)
		nl.baseRecord.UsedFields++
	}
	return nl
}

func (l *logger) Float64(name string, f float64) *logger {
	nl := l
	if nl.baseRecord.UsedFields < maxFields {
		i := nl.baseRecord.UsedFields
		nl.baseRecord.SetField(i, name, f)
		nl.baseRecord.UsedFields++
	}
	return nl
}

func (l *logger) Bool(name string, b bool) *logger {
	nl := l
	if nl.baseRecord.UsedFields < maxFields {
		i := nl.baseRecord.UsedFields
		nl.baseRecord.SetField(i, name, b)
		nl.baseRecord.UsedFields++
	}
	return nl
}

func (l *logger) Any(name string, v any) *logger {
	nl := l
	if nl.baseRecord.UsedFields < maxFields {
		i := nl.baseRecord.UsedFields
		nl.baseRecord.SetField(i, name, v)
		nl.baseRecord.UsedFields++
	}
	return nl
}

func (l *logger) WithError(err error) Logger {
	if err == nil {
		return l
	}
	return l.WithField("error", err.Error())
}
