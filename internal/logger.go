package internal

// A Logger represents a logger.
type Logger interface {
	// Printf logs a message with timestamp but without level and module.
	Printf(string, ...any)
	// Print logs a message with timestamp but without level and module.
	Print(...any)

	// Trace logs a message at trace level.
	Trace(...any)
	// Tracef logs a message at trace level.
	Tracef(string, ...any)
	// Debug logs a message at debug level.
	Debug(...any)
	// Debugf logs a message at debug level.
	Debugf(string, ...any)
	// Warn logs a message at warn level.
	Warn(...any)
	// Warnf logs a message at warn level.
	Warnf(string, ...any)
	// Info logs a message at info level.
	Info(...any)
	// Infof logs a message at info level.
	Infof(string, ...any)
	// Error logs a message at error level.
	Error(...any)
	// Errorf logs a message at error level.
	Errorf(string, ...any)

	// WriteRawString writes a raw message with module.
	WriteRawString(string)

	// WithTraceId returns a new logger with trace id.
	WithTraceId(string) Logger
}
