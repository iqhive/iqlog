package iqlog

const invalidTimestampSecond int64 = -1 << 63

// Event is a single-use structured log event obtained from a Logger; the
// zero value is not usable. It is not safe for concurrent use. Msg, Msgs,
// Msgf, and Discard consume the event and recycle it, so the handle must not
// be used afterwards: a later call through it can act on a different event
// that has since reused the same storage.
type Event struct {
	logger   *Logger
	config   *loggerConfig
	jsonMode bool
	output   []byte
	// prefixLen marks the end of the package-generated console envelope
	// (timestamp, level, caller). Everything after it is caller-supplied and
	// gets control bytes escaped before the record is written.
	prefixLen       int
	includeTime     bool
	color           bool
	captureCaller   int
	disabled        bool
	exitAfterWrite  bool
	panicAfterWrite bool
	panicMessage    string
	callerData      callerData
	consumed        bool
	buildErr        error
	level           Level
	// The fields below survive pooling on purpose: they are a cache keyed on
	// the clock, not on the record, and every Event that comes out of the
	// pool is better off with it warm.
	//
	// timePrefix holds `{"time":"` followed by the UTC date-time of
	// timeSecond, so a record in the same second as the last one copies 28
	// bytes and formats six microsecond digits. invalidTimestampSecond means
	// the prefix has never been built.
	timeSecond int64
	timePrefix [jsonTimePrefixLen]byte
}

func acquireEvent(l *Logger, cfg *loggerConfig) *Event {
	e := eventPool.Get().(*Event)
	if cap(e.output) == 0 {
		e.output = acquireEventBuffer()
	}
	e.logger = l
	e.config = cfg
	e.jsonMode = false
	e.output = e.output[:0]
	e.prefixLen = 0
	e.includeTime = false
	e.color = false
	e.captureCaller = 0
	e.disabled = false
	e.exitAfterWrite = false
	e.panicAfterWrite = false
	e.panicMessage = ""
	e.callerData.callerFuncLen = 0
	e.callerData.callerFileLen = 0
	e.consumed = false
	e.buildErr = nil
	e.level = LevelUnknown
	return e
}

func (e *Event) release(releaseBuffer bool) {
	buf := e.output
	if !releaseBuffer || cap(buf) > maxPooledCapacity {
		buf = nil
	}
	e.logger = nil
	e.config = nil
	e.jsonMode = false
	e.output = buf
	e.prefixLen = 0
	e.includeTime = false
	e.color = false
	e.captureCaller = 0
	e.disabled = false
	e.exitAfterWrite = false
	e.panicAfterWrite = false
	e.panicMessage = ""
	e.callerData.callerFuncLen = 0
	e.callerData.callerFileLen = 0
	e.consumed = true
	e.buildErr = nil
	eventPool.Put(e)
}
