package iqlog

const maxFields = 8

type LogFields struct {
	inlineSize    int
	keys          [maxFields]string
	values        [maxFields]any
	usedInline    int
	dynamicKeys   []string
	dynamicValues []any
}

func NewLogFields() LogFields {
	return LogFields{
		inlineSize: maxFields,
	}
}

func (lf *LogFields) AddField(key string, val any) {
	if lf.usedInline < lf.inlineSize {
		lf.keys[lf.usedInline] = key
		lf.values[lf.usedInline] = val
		lf.usedInline++
		return
	}
	lf.dynamicKeys = append(lf.dynamicKeys, key)
	lf.dynamicValues = append(lf.dynamicValues, val)
}

func (lf LogFields) Copy() LogFields {
	newLF := lf
	if len(lf.dynamicKeys) > 0 {
		newLF.dynamicKeys = append([]string(nil), lf.dynamicKeys...)
	}
	if len(lf.dynamicValues) > 0 {
		newLF.dynamicValues = append([]any(nil), lf.dynamicValues...)
	}
	return newLF
}

func (lf *LogFields) MergeAll(other LogFields) {
	for i := 0; i < other.usedInline; i++ {
		lf.AddField(other.keys[i], other.values[i])
	}
	for i := range other.dynamicKeys {
		lf.AddField(other.dynamicKeys[i], other.dynamicValues[i])
	}
}

func (lf *LogFields) ForEach(fn func(key string, val any)) {
	for i := 0; i < lf.usedInline; i++ {
		fn(lf.keys[i], lf.values[i])
	}
	for i := range lf.dynamicKeys {
		fn(lf.dynamicKeys[i], lf.dynamicValues[i])
	}
}
