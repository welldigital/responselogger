package responselogger

import (
	"bytes"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

// Logger defines how HTTP requests are logged, e.g. to the console, or in JSON format (see JSONLogger).
type Logger func(url *url.URL, status int, len int64, d time.Duration)

// JSONLogger uses logrus to log the HTTP request in JSON format to os.Stderr.
func JSONLogger(url *url.URL, status int, len int64, d time.Duration) {
	os.Stderr.WriteString(JSONLogMessage(time.Now, url, status, len, d))
}

var jsonEscapesMap = map[rune]string{
	0x0022: `\"`,
	0x005C: `\\`,
	0x0008: `\b`,
	0x000C: `\f`,
	0x000A: `\n`,
	0x000D: `\r`,
	0x0009: `\n`,
}

func jsonEscape(s string) string {
	b := bytes.NewBufferString("")
	for _, r := range s {
		// Skip control chars, they're not valid in URLs either.
		if r >= 0x0000 && r <= 0x001F {
			continue
		}
		// Replace others with escaped values.
		if replacement, ok := jsonEscapesMap[r]; ok {
			b.WriteString(replacement)
			continue
		}
		// Use the character.
		b.WriteRune(r)
	}
	return b.String()
}

// JSONLogMessage formats a log message to JSON.
func JSONLogMessage(now func() time.Time, u *url.URL, status int, len int64, d time.Duration) string {
	c := "http_" + strconv.Itoa(status/100) + "xx"
	return `{` +
		`"time":"` + now().UTC().Format(time.RFC3339) + `",` +
		`"src":"rl",` +
		`"status":` + strconv.Itoa(status) + `,` +
		`"` + c + `":1,` +
		`"len":` + strconv.FormatInt(len, 10) + `,` +
		`"ms":` + strconv.FormatInt(d.Nanoseconds()/1000000, 10) + `,` +
		`"path":"` + jsonEscape(u.Path) + `"}` + "\n"
}

// Handler provides a way to log HTTP requests - the status code, http category, size and duration.
type Handler struct {
	Next   http.Handler
	Logger Logger
	Skip   func(r *http.Request) bool
}

// NewHandler creates a new responselogger.Handler with default JSON logger which skips logging '/health' URLs.
func NewHandler(next http.Handler) Handler {
	return Handler{
		Next:   next,
		Logger: JSONLogger,
		Skip:   SkipHealthEndpoint,
	}
}

// SkipHealthEndpoint rejects logging the /health URL.
func SkipHealthEndpoint(r *http.Request) bool {
	return r.URL.Path == "/health"
}

// ServeHTTP handles the HTTP request, keeping track of the status code used.
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Skip(r) {
		h.Next.ServeHTTP(w, r)
		return
	}
	var written int64
	var status = -1

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
			w.WriteHeader(code)
		},
	}

	start := time.Now()
	h.Next.ServeHTTP(wp, r)
	duration := time.Now().Sub(start)

	// Use default status.
	if status == -1 {
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
