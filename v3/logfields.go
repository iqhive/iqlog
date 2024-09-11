package iqlog

import "time"

// We allow a fixed maximum of fields so we can try avoid an alloc
const maxFields = 8
const maxStringLen = 256

// LogField is used to store each field in a (pre-alloc'd) array
type LogField struct {
	Key    [maxStringLen]byte // fixed-size key
	KeyLen int                // length of the key
	VStr   [maxStringLen]byte // fixed-size string representation
	VLen   int                // length of the "string" portion
	Quote  bool               // whether to quote the value
	Used   bool               // whether this slot is in use
}

type LogRecord struct {
	Time       time.Time
	Level      Level
	Message    [maxStringLen]byte // fixed-size message buffer
	MsgLen     int                // length of message
	Fields     [maxFields]LogField
	UsedFields int
}

type LogFields struct {
	inlineSize    int
	keys          [8]string
	values        [8]any
	usedInline    int
	dynamicKeys   []string
	dynamicValues []any
}

func NewLogFields() LogFields {
	return LogFields{
		inlineSize: 8,
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
