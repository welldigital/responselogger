package responselogger

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

// Logger defines how HTTP requests are logged, e.g. to the console, or in JSON format (see JSONLogger).
type Logger func(url *url.URL, status int, len int64, d time.Duration)

// JSONLogger uses logrus to log the HTTP request in JSON format to os.Stderr.
func JSONLogger(url *url.URL, status int, len int64, d time.Duration) {
	os.Stderr.WriteString(JSONLogMessage(url, status, len, d))
}

// JSONLogMessage formats a log message to JSON.
func JSONLogMessage(url *url.URL, status int, len int64, d time.Duration) string {
	s := status / 100

	return fmt.Sprintf(`{ "src": "rl", "status": %d, "%s": 1, "len": %d, "ms": %d, "path": "%s" }`,
		status,
		fmt.Sprintf("http_%dxx", s),
		len,
		d.Nanoseconds()/1000000,
		url.Path)
}

// Handler provides a way to log HTTP requests - the status code, http category, size and duration.
type Handler struct {
	Next   http.Handler
	Logger Logger
}

// ServeHTTP handles the HTTP request, keeping track of the status code used.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var written int64
	var status int
	statusWritten := false

	wp := writerProxy{
		h: func() http.Header {
			return w.Header()
		},
		w: func(bytes []byte) (int, error) {
			bw, err := w.Write(bytes)
			written += int64(bw)
			return bw, err
		},
		wh: func(code int) {
			status = code
			statusWritten = true
			w.WriteHeader(code)
		},
	}

	start := time.Now()
	h.Next.ServeHTTP(wp, r)
	duration := time.Now().Sub(start)

	// Use default status.
	if !statusWritten {
		status = 200
	}

	h.Logger(r.URL, status, written, duration)
}

type writerProxy struct {
	h  func() http.Header
	w  func(bytes []byte) (int, error)
	wh func(status int)
}

func (wp writerProxy) Header() http.Header {
	return wp.h()
}

func (wp writerProxy) Write(bytes []byte) (int, error) {
	return wp.w(bytes)
}

func (wp writerProxy) WriteHeader(status int) {
	wp.wh(status)
}
