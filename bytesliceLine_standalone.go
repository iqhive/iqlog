package iqlog

// Event is a single-use structured log event. It is not safe for concurrent
// use. Msg, Msgs, or Msgf consumes the event.
type Event struct {
	logger          *Logger
	config          *loggerConfig
	jsonMode        bool
	output          []byte
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
}

func acquireEvent(l *Logger, cfg *loggerConfig) *Event {
	e := eventPool.Get().(*Event)
	buf := e.output
	if cap(buf) == 0 {
		buf = acquireEventBuffer()
	}
	*e = Event{logger: l, config: cfg, output: buf[:0]}
	return e
}

func (e *Event) release(releaseBuffer bool) {
	buf := e.output
	if !releaseBuffer || cap(buf) > maxPooledCapacity {
		buf = nil
	}
	*e = Event{output: buf, consumed: true}
	eventPool.Put(e)
}
