package iqlog

import "time"

func ansiColourPrefix(level Level) []byte {
	switch level {
	case LevelInfo:
		return []byte("\x1b[34mINFO\x1b[0m ")
	case LevelDebug, LevelTrace:
		return []byte("\x1b[37mDBUG\x1b[0m ")
	case LevelWarn:
		return []byte("\x1b[33mWARN\x1b[0m ")
	case LevelError:
		return []byte("\x1b[31mERRR\x1b[0m ")
	case LevelFatal:
		return []byte("\x1b[31mFATL\x1b[0m ")
	case LevelPanic:
		return []byte("\x1b[31mPANC\x1b[0m ")
	default:
		return []byte("\x1b[34m????\x1b[0m ")
	}
}

// levelPrefix returns a constant []byte for each level (no allocations).
func levelPrefix(level Level) []byte {
	switch level {
	case LevelInfo:
		return []byte("INFO ")
	case LevelDebug, LevelTrace:
		return []byte("DBUG ")
	case LevelWarn:
		return []byte("WARN ")
	case LevelError:
		return []byte("ERRR ")
	case LevelFatal:
		return []byte("FATL ")
	case LevelPanic:
		return []byte("PANC ")
	default:
		return []byte("???? ")
	}
}

func safeStringCopy(dst *[maxStringLen]byte, s string) int {
	n := len(s)
	if n > maxStringLen {
		n = maxStringLen
	}
	copy(dst[:n], s)
	if n < maxStringLen {
		dst[n] = 0
	}
	return n
}

// appendTimeRFC3339Micro encodes t as:  yyyy-mm-ddThh:mm:ss.uuuuuu
// returning how many bytes were written.  This avoids any string
// allocations from time.Time.Format.
func appendTimeRFC3339Micro(t time.Time, dst []byte) int {
	year, month, day := t.Date()
	hour, min, sec := t.Clock()
	usec := t.Nanosecond() / 1000

	pos := 0
	pos += appendIntBytes(dst[pos:], int64(year), 4)
	dst[pos] = '-'
	pos++
	pos += appendIntBytes(dst[pos:], int64(month), 2)
	dst[pos] = '-'
	pos++
	pos += appendIntBytes(dst[pos:], int64(day), 2)
	dst[pos] = 'T'
	pos++
	pos += appendIntBytes(dst[pos:], int64(hour), 2)
	dst[pos] = ':'
	pos++
	pos += appendIntBytes(dst[pos:], int64(min), 2)
	dst[pos] = ':'
	pos++
	pos += appendIntBytes(dst[pos:], int64(sec), 2)
	dst[pos] = '.'
	pos++
	pos += appendIntBytes(dst[pos:], int64(usec), 6)
	return pos
}

func writeIntDecimal(dst []byte, i int64) int {
	neg := (i < 0)
	if neg {
		i = -i
	}
	var tmp [20]byte
	pos := len(tmp)

	if i == 0 {
		pos--
		tmp[pos] = '0'
	} else {
		for i > 0 {
			pos--
			tmp[pos] = byte('0' + (i % 10))
			i /= 10
		}
	}
	if neg {
		pos--
		tmp[pos] = '-'
	}

	n := copy(dst, tmp[pos:])
	if n < len(dst) {
		dst[n] = 0
	}
	return n
}

func fastFloatFill(dst []byte, f float64, decimals int) int {
	neg := (f < 0)
	if neg {
		f = -f
	}

	intPart := int64(f)
	written := 0

	if neg {
		if written < len(dst) {
			dst[written] = '-'
			written++
		}
	}

	written += writeIntDecimal(dst[written:], intPart)
	frac := f - float64(intPart)
	if frac == 0.0 {
		return written
	}
	if written < len(dst) {
		dst[written] = '.'
		written++
	}
	for i := 0; i < decimals; i++ {
		frac *= 10
		d := int64(frac)
		if written < len(dst) {
			dst[written] = byte('0' + d)
			written++
		}
		frac -= float64(d)
		if frac == 0.0 {
			break
		}
	}
	if written < len(dst) {
		dst[written] = 0
	}
	return written
}
