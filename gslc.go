package gslc

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/hashamali/gsl"
	"github.com/rs/zerolog"
)

// Middleware will return a new Fiber middleware for logging.
func Middleware(logger gsl.Log) func(http.Handler) http.Handler {
	return middleware.RequestLogger(&httpLogger{Logger: logger})
}

type httpLogger struct {
	Logger gsl.Log
}

func (l httpLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	rid := r.Context().Value(middleware.RequestIDKey).(string)
	scheme := r.URL.Scheme
	if scheme == "" {
		scheme = "http"
	}

	return &log{
		Logger:   l.Logger,
		ID:       rid,
		RemoteIP: r.RemoteAddr,
		Method:   r.Method,
		Host:     r.Host,
		Path:     r.URL.Path,
		Protocol: scheme,
	}
}

type log struct {
	Logger     gsl.Log
	ID         string
	RemoteIP   string
	Host       string
	Method     string
	Path       string
	Protocol   string
	Bytes      int
	StatusCode int
	Latency    float64
	Error      error
	Stack      []byte
}

func (l *log) MarshalZerologObject(zle *zerolog.Event) {
	zle.Str("id", l.ID)
	zle.Str("remote_ip", l.RemoteIP)
	zle.Str("host", l.Host)
	zle.Str("method", l.Method)
	zle.Str("path", l.Path)
	zle.Str("protocol", l.Protocol)
	zle.Int("status_code", l.StatusCode)
	zle.Float64("latency", l.Latency)

	if l.Error != nil {
		zle.Err(l.Error)
	}
}

func (l *log) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	l.StatusCode = status
	l.Bytes = bytes
	l.Latency = float64(elapsed) / float64(time.Millisecond)
	l.Logger.Infow(l, "")
}

func (l *log) Panic(v interface{}, stack []byte) {
	err := v.(error)
	if err != nil {
		l.Error = err
	} else {
		l.Error = errors.New("unknown error")
	}

	l.Stack = stack
	l.Logger.Errorw(l, "")
}
