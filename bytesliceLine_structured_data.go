package iqlog

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

func (bsl *Event) appendJSONKey(name string) {
	bsl.output = append(bsl.output, ',', '"')
	if bsl.config.escapeFieldNames {
		bsl.output = appendJSONEscaped(bsl.output, name)
	} else {
		bsl.output = append(bsl.output, name...)
	}
	bsl.output = append(bsl.output, '"', ':')
}

func eventFieldName(name string) string {
	switch name {
	case "time", "level", "message", "func", "file":
		return "field_" + name
	default:
		return name
	}
}

func (bsl *Event) Int(name string, val int) *Event {
	if bsl == nil {
		return bsl
	}
	name = eventFieldName(name)
	if bsl.jsonMode {
		bsl.appendJSONKey(name)
		bsl.output, _ = appendIntDecimal(bsl.output, int64(val))
	} else {
		bsl.output = append(bsl.output, name...)
		bsl.output = append(bsl.output, '=')
		bsl.output, _ = appendIntDecimal(bsl.output, int64(val))
		bsl.output = append(bsl.output, ' ')
	}
	return bsl
}

func (bsl *Event) Int64(name string, val int64) *Event {
	if bsl == nil {
		return bsl
	}
	name = eventFieldName(name)
	if bsl.jsonMode {
		bsl.appendJSONKey(name)
		bsl.output, _ = appendIntDecimal(bsl.output, int64(val))
	} else {
		bsl.output = append(bsl.output, name...)
		bsl.output = append(bsl.output, '=')
		bsl.output, _ = appendIntDecimal(bsl.output, int64(val))
		bsl.output = append(bsl.output, ' ')
	}
	return bsl
}

func (bsl *Event) Str(name string, s string) *Event {
	if bsl == nil {
		return bsl
	}
	name = eventFieldName(name)
	if bsl.jsonMode {
		bsl.appendJSONKey(name)
		bsl.output = append(bsl.output, '"')
		bsl.output = appendJSONEscaped(bsl.output, s)
		bsl.output = append(bsl.output, '"')
	} else {
		bsl.output = append(bsl.output, name...)
		bsl.output = append(bsl.output, '=')
		bsl.output = append(bsl.output, s...)
		bsl.output = append(bsl.output, ' ')
	}
	return bsl
}

func (bsl *Event) Float32(name string, f float32) *Event {
	if bsl == nil {
		return bsl
	}
	name = eventFieldName(name)
	if bsl.jsonMode {
		bsl.appendJSONKey(name)
		bsl.output = appendJSONFloat(bsl.output, float64(f), 32)
	} else {
		bsl.output = append(bsl.output, name...)
		bsl.output = append(bsl.output, '=')
		bsl.output = strconv.AppendFloat(bsl.output, float64(f), 'f', -1, 32)
		bsl.output = append(bsl.output, ' ')
	}
	return bsl
}

func (bsl *Event) Float64(name string, f float64) *Event {
	if bsl == nil {
		return bsl
	}
	name = eventFieldName(name)
	if bsl.jsonMode {
		bsl.appendJSONKey(name)
		bsl.output = appendJSONFloat(bsl.output, f, 64)
	} else {
		bsl.output = append(bsl.output, name...)
		bsl.output = append(bsl.output, '=')
		bsl.output = strconv.AppendFloat(bsl.output, f, 'f', -1, 64)
		bsl.output = append(bsl.output, ' ')
	}
	return bsl
}

func (bsl *Event) Bool(name string, b bool) *Event {
	if bsl == nil {
		return bsl
	}
	name = eventFieldName(name)
	if bsl.jsonMode {
		bsl.appendJSONKey(name)
		if b {
			bsl.output = append(bsl.output, "true"...)
		} else {
			bsl.output = append(bsl.output, "false"...)
		}
	} else {
		if b {
			bsl.output = append(bsl.output, name...)
			bsl.output = append(bsl.output, "=true "...)
		} else {
			bsl.output = append(bsl.output, name...)
			bsl.output = append(bsl.output, "=false "...)
		}
	}
	return bsl
}

func (bsl *Event) Any(name string, v any) *Event {
	if bsl == nil {
		return bsl
	}
	switch value := v.(type) {
	case nil:
		name = eventFieldName(name)
		if bsl.jsonMode {
			bsl.appendJSONKey(name)
			bsl.output = append(bsl.output, "null"...)
		} else {
			bsl.output = append(bsl.output, name...)
			bsl.output = append(bsl.output, "=null "...)
		}
		return bsl
	case string:
		return bsl.Str(name, value)
	case bool:
		return bsl.Bool(name, value)
	case int:
		return bsl.Int(name, value)
	case int8:
		return bsl.Int64(name, int64(value))
	case int16:
		return bsl.Int64(name, int64(value))
	case int32:
		return bsl.Int64(name, int64(value))
	case int64:
		return bsl.Int64(name, value)
	case uint:
		return bsl.Uint(name, value)
	case uint8:
		return bsl.Uint64(name, uint64(value))
	case uint16:
		return bsl.Uint64(name, uint64(value))
	case uint32:
		return bsl.Uint64(name, uint64(value))
	case uint64:
		return bsl.Uint64(name, value)
	case float32:
		return bsl.Float32(name, value)
	case float64:
		return bsl.Float64(name, value)
	case error:
		return bsl.Str(name, value.Error())
	case time.Time:
		return bsl.Time(name, value)
	case time.Duration:
		return bsl.Duration(name, value)
	case json.RawMessage:
		return bsl.RawJSON(name, value)
	case []byte:
		return bsl.Bytes(name, value)
	}
	name = eventFieldName(name)
	if bsl.jsonMode {
		bsl.appendJSONKey(name)
		encoded, err := json.Marshal(v)
		if err == nil {
			bsl.output = append(bsl.output, encoded...)
		} else {
			bsl.output = append(bsl.output, '"')
			bsl.output = appendJSONEscaped(bsl.output, fmt.Sprintf("%v", v))
			bsl.output = append(bsl.output, '"')
		}
	} else {
		bsl.output = append(bsl.output, []byte(name+"="+fmt.Sprintf("%v", v)+" ")...)
	}
	return bsl
}

func (e *Event) Uint(name string, value uint) *Event {
	return e.Uint64(name, uint64(value))
}

func (e *Event) Uint64(name string, value uint64) *Event {
	if e == nil {
		return e
	}
	name = eventFieldName(name)
	if e.jsonMode {
		e.appendJSONKey(name)
	} else {
		e.output = append(e.output, name...)
		e.output = append(e.output, '=')
	}
	e.output = strconv.AppendUint(e.output, value, 10)
	if !e.jsonMode {
		e.output = append(e.output, ' ')
	}
	return e
}

func (e *Event) Err(err error) *Event {
	if e == nil || err == nil {
		return e
	}
	return e.Str("error", err.Error())
}

func (e *Event) Stringer(name string, value fmt.Stringer) *Event {
	if e == nil || value == nil {
		return e
	}
	return e.Str(name, value.String())
}

func (e *Event) Time(name string, value time.Time) *Event {
	if e == nil {
		return nil
	}
	name = eventFieldName(name)
	if e.jsonMode {
		e.appendJSONKey(name)
		e.output = append(e.output, '"')
		e.output = value.AppendFormat(e.output, time.RFC3339Nano)
		e.output = append(e.output, '"')
	} else {
		e.output = append(e.output, name...)
		e.output = append(e.output, '=')
		e.output = value.AppendFormat(e.output, time.RFC3339Nano)
		e.output = append(e.output, ' ')
	}
	return e
}

func (e *Event) Duration(name string, value time.Duration) *Event {
	if e == nil {
		return nil
	}
	return e.Str(name, value.String())
}

func (e *Event) Bytes(name string, value []byte) *Event {
	if e == nil {
		return nil
	}
	name = eventFieldName(name)
	if e.jsonMode {
		e.appendJSONKey(name)
		e.output = append(e.output, '"')
		e.output = base64.StdEncoding.AppendEncode(e.output, value)
		e.output = append(e.output, '"')
	} else {
		e.output = append(e.output, name...)
		e.output = append(e.output, '=')
		e.output = base64.StdEncoding.AppendEncode(e.output, value)
		e.output = append(e.output, ' ')
	}
	return e
}

func (e *Event) RawJSON(name string, value []byte) *Event {
	if e == nil {
		return e
	}
	name = eventFieldName(name)
	if !json.Valid(value) {
		e.buildErr = fmt.Errorf("iqlog: invalid raw JSON for field %q", name)
		value = []byte("null")
	}
	if e.jsonMode {
		e.appendJSONKey(name)
		e.output = append(e.output, value...)
	} else {
		e.output = append(e.output, name...)
		e.output = append(e.output, '=')
		e.output = append(e.output, value...)
		e.output = append(e.output, ' ')
	}
	return e
}

// BuildError reports a field-construction error such as invalid raw JSON.
func (e *Event) BuildError() error {
	if e == nil {
		return nil
	}
	return e.buildErr
}
