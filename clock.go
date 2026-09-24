package iqlog

import "time"

// The two fixed-width timestamps, JSONTimeUTC and the default console layout,
// resolve to the microsecond, so a record stamped by either needs exactly one
// reading of the wall clock and nothing from the monotonic clock that
// time.Now also reads. wallClock, in clock_linux_amd64.go and clock_other.go,
// is that single reading, and the event path takes it only when the
// configuration's fastClock is set: Config.Now is time.Now itself and the
// encoder is one of the two. Every other combination calls Config.Now for
// every record exactly as configured, so a nanosecond layout keeps its digits
// and a caller's clock is honoured as given.

// twoDigits is every value 00-99 as two ASCII digits, so a two-digit field
// costs one multiply and two loads instead of a divide and a modulo.
const twoDigits = "00010203040506070809" +
	"10111213141516171819" +
	"20212223242526272829" +
	"30313233343536373839" +
	"40414243444546474849" +
	"50515253545556575859" +
	"60616263646566676869" +
	"70717273747576777879" +
	"80818283848586878889" +
	"90919293949596979899"

const (
	// jsonTimePrefixLen is len(`{"time":"2006-01-02T15:04:05`), the part of
	// the fixed-width JSON timestamp that only changes once a second.
	jsonTimePrefixLen = 28
	// jsonTimestampLen adds `.000000Z",` to the prefix.
	jsonTimestampLen = jsonTimePrefixLen + 10
)

// writeDefaultJSONTimestamp appends {"time":"YYYY-MM-DDTHH:MM:SS.uuuuuuZ",
// for a Unix second and nanosecond, which must lie in [0, 1e9). The field
// carries microseconds, so the nanoseconds are truncated. This is the entry
// for a time.Time from a caller's clock or an adapter's record; the default
// clock already reads microseconds and enters writeJSONTimestampMicros
// directly.
func (e *Event) writeDefaultJSONTimestamp(sec, nsec int64) {
	e.writeJSONTimestampMicros(sec, int64(uint64(nsec)/1000))
}

// writeJSONTimestampMicros appends the fixed-width UTC timestamp for a Unix
// second and microsecond, which must lie in [0, 1e6). The date-time prefix is
// cached on the Event and rebuilt only when the second changes, so the common
// record costs one 28-byte copy and six microsecond digits, written straight
// into the output rather than through several appends.
func (e *Event) writeJSONTimestampMicros(sec, usec int64) {
	if sec != e.timeSecond && !e.cacheJSONTimePrefix(sec) {
		// A year outside 0-9999 does not fit the fixed width; the general
		// formatter widens or signs it, and the cache keeps its last good
		// second.
		e.output = appendJSONTimestampSlow(e.output, time.Unix(sec, usec*1000).UTC())
		return
	}
	n := len(e.output)
	if cap(e.output)-n < jsonTimestampLen {
		// the pooled buffer always has room; this is the fresh-buffer case
		e.output = append(e.output, make([]byte, jsonTimestampLen)...)
	} else {
		e.output = e.output[:n+jsonTimestampLen]
	}
	out := e.output[n : n+jsonTimestampLen : n+jsonTimestampLen]
	// an array assignment compiles to a few register moves where copy would
	// call memmove for a source it cannot prove disjoint
	*(*[jsonTimePrefixLen]byte)(out) = e.timePrefix
	// unsigned so the divisions by constants need no sign correction; two
	// divisions and two multiply-subtracts give the three digit pairs
	u := uint32(usec)
	hi := u / 10000
	rem := u - hi*10000
	mid := rem / 100
	lo := rem - mid*100
	hi *= 2
	mid *= 2
	lo *= 2
	out[jsonTimePrefixLen] = '.'
	out[jsonTimePrefixLen+1] = twoDigits[hi]
	out[jsonTimePrefixLen+2] = twoDigits[hi+1]
	out[jsonTimePrefixLen+3] = twoDigits[mid]
	out[jsonTimePrefixLen+4] = twoDigits[mid+1]
	out[jsonTimePrefixLen+5] = twoDigits[lo]
	out[jsonTimePrefixLen+6] = twoDigits[lo+1]
	out[jsonTimePrefixLen+7] = 'Z'
	out[jsonTimePrefixLen+8] = '"'
	out[jsonTimePrefixLen+9] = ','
}

// cacheJSONTimePrefix rebuilds timePrefix for the Unix second and records
// it in timeSecond, or reports false, leaving the cache alone, when the UTC
// year does not fit four digits.
func (e *Event) cacheJSONTimePrefix(sec int64) bool {
	utc := time.Unix(sec, 0).UTC()
	year, month, day := utc.Date()
	if uint(year) > 9999 {
		return false
	}
	hour, minute, second := utc.Clock()
	p := &e.timePrefix
	copy(p[:9], `{"time":"`)
	century, yy := year/100*2, year%100*2
	p[9], p[10], p[11], p[12] = twoDigits[century], twoDigits[century+1], twoDigits[yy], twoDigits[yy+1]
	p[13] = '-'
	mm := int(month) * 2
	p[14], p[15] = twoDigits[mm], twoDigits[mm+1]
	p[16] = '-'
	dd := day * 2
	p[17], p[18] = twoDigits[dd], twoDigits[dd+1]
	p[19] = 'T'
	hh := hour * 2
	p[20], p[21] = twoDigits[hh], twoDigits[hh+1]
	p[22] = ':'
	mi := minute * 2
	p[23], p[24] = twoDigits[mi], twoDigits[mi+1]
	p[25] = ':'
	ss := second * 2
	p[26], p[27] = twoDigits[ss], twoDigits[ss+1]
	e.timeSecond = sec
	return true
}
