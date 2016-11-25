// Package accesslog provides http.Handler wrapper that logs requests.
package accesslog

import "net/http"

// LogFunc is a custom function doing logging. This function should normally
// capture request and return a closure writing log entry when called with
// response code as a single argument. Function should not modify request. See
// example for implementation detail.
type LogFunc func(r *http.Request) func(responseCode int)

// Logger describes method used to print entries by handler returned by
// WithLog() function. Standard library *log.Logger implements this interface.
type Logger interface {
	Printf(format string, v ...interface{})
}

// WithCustomLog wraps original handler and returns another one that logs
// requests using provided LogFunc on first call to WriteHeader() or Write() on
// http.ResponseWriter by original handler, whichever comes first.
//
// ResponseWriter original handler sees does not implement http.Hijacker
// interface.
func WithCustomLog(h http.Handler, fn LogFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(&loggingResponseWriter{w: w, log: fn(r)}, r)
	})
}

// WithLog wraps handler and returns another one logging each request in
// a simple format. Log entry is written to Logger on either WriteHeader() or
// first Write() calls to http.ResponseWriter, whichever comes first.
func WithLog(h http.Handler, l Logger) http.Handler {
	fn := func(r *http.Request) func(int) {
		return func(code int) {
			l.Printf("%s %s %d %q", r.Method, r.URL, code, r.UserAgent())
		}
	}
	return WithCustomLog(h, fn)
}

// loggingResponseWriter implements http.ResponseWriter interface that wraps
// another ResponseWriter and calls log function with response code on first
// WriteHeader of Write call.
type loggingResponseWriter struct {
	w   http.ResponseWriter
	log func(int)
}

func (rw *loggingResponseWriter) Header() http.Header { return rw.w.Header() }
func (rw *loggingResponseWriter) WriteHeader(code int) {
	if rw.log != nil {
		rw.log(code)
		rw.log = nil
	}
	rw.w.WriteHeader(code)
}
func (rw *loggingResponseWriter) Write(b []byte) (int, error) {
	if rw.log != nil {
		// rw.log != nil means WriteHeader was not called, so implicitly
		// assume 200 OK
		rw.log(http.StatusOK)
		rw.log = nil
	}
	return rw.w.Write(b)
}
