package responselogger

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

var httpResponseBody = "<html>\n<head>\n<title>Example</title>\n</head>\n<body>\nExample\n</body></html>"

func TestHandlerWithoutDelay(t *testing.T) {
	tests := []struct {
		name                  string
		handler               http.HandlerFunc
		expectedStatus        int
		expectedLength        int
		expectedBody          string
		r                     *http.Request
		expectedMessageLogged bool
	}{
		{
			name: ">= 0,< 100",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(50)
				w.Write([]byte(httpResponseBody))
			},
			expectedStatus: 50,
			expectedLength: len(httpResponseBody),
			expectedBody:   httpResponseBody,
			r:              httptest.NewRequest(http.MethodGet, "/index.html", nil),
			expectedMessageLogged: true,
		},
		{
			name: ">= 100, < 200",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(101)
				w.Write([]byte(httpResponseBody))
			},
			expectedStatus: 101,
			expectedLength: len(httpResponseBody),
			expectedBody:   httpResponseBody,
			r:              httptest.NewRequest(http.MethodGet, "/index.html", nil),
			expectedMessageLogged: true,
		},
		{
			name: "200 (default)",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(httpResponseBody))
			},
			expectedStatus: 200,
			expectedLength: len(httpResponseBody),
			expectedBody:   httpResponseBody,
			r:              httptest.NewRequest(http.MethodGet, "/index.html", nil),
			expectedMessageLogged: true,
		},
		{
			name: "200 (specific)",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				w.Write([]byte(httpResponseBody))
			},
			expectedStatus: 200,
			expectedLength: len(httpResponseBody),
			expectedBody:   httpResponseBody,
			r:              httptest.NewRequest(http.MethodGet, "/index.html", nil),
			expectedMessageLogged: true,
		},
		{
			name: ">= 200, < 300",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(201)
				w.Write([]byte(httpResponseBody))
			},
			expectedStatus: 201,
			expectedLength: len(httpResponseBody),
			expectedBody:   httpResponseBody,
			r:              httptest.NewRequest(http.MethodGet, "/index.html", nil),
			expectedMessageLogged: true,
		},
		{
			name: "redirect",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/other", http.StatusMovedPermanently)
			},
			expectedStatus: 301,
			expectedLength: len("<a href=\"/other\">Moved Permanently</a>.\n\n"),
			expectedBody:   "<a href=\"/other\">Moved Permanently</a>.\n\n",
			r:              httptest.NewRequest(http.MethodGet, "/index.html", nil),
			expectedMessageLogged: true,
		},
		{
			name: "http 500 error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "error!", 500)
			},
			expectedStatus: 500,
			expectedLength: len("error!\n"),
			expectedBody:   "error!\n",
			r:              httptest.NewRequest(http.MethodGet, "/index.html", nil),
			expectedMessageLogged: true,
		},
		{
			name: "skip health check",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(201)
				w.Write([]byte(httpResponseBody))
			},
			expectedStatus: 201,
			expectedBody:   httpResponseBody,
			r:              httptest.NewRequest(http.MethodGet, "/health", nil),
			expectedMessageLogged: false,
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()

		var loggedURL string
		var loggedStatus int
		var loggedLength int64
		var messageLogged bool

		h := Handler{
			Next: test.handler,
			Logger: func(url *url.URL, status int, len int64, d time.Duration) {
				loggedURL = url.String()
				loggedStatus = status
				loggedLength = len
				messageLogged = true
			},
			Skip: SkipHealthEndpoint,
		}
		h.ServeHTTP(w, test.r)

		if test.expectedMessageLogged != messageLogged {
			t.Fatalf("%s: expected messageLogged to be %v, but was %v", test.name, test.expectedMessageLogged, messageLogged)
		}

		if messageLogged {
			if test.r.URL.String() != loggedURL {
				t.Errorf("%s: expected URL %v to be logged, but got %v", test.name, test.r.URL.String(), loggedURL)
			}
			if test.expectedStatus != loggedStatus {
				t.Errorf("%s: expected status %d to be logged, but got %d", test.name, test.expectedStatus, loggedStatus)
			}
			if int64(test.expectedLength) != loggedLength {
				t.Errorf("%s: expected length %d to be logged, but got %d", test.name, test.expectedLength, loggedLength)
			}
		}

		// The data sent to the ResponseWriter should always be correct, regardless of whether we log.
		if test.expectedBody != w.Body.String() {
			t.Errorf("%s: expected body of '%s', got '%s'", test.name, test.expectedBody, w.Body.String())
		}
		if test.expectedStatus != w.Code {
			t.Errorf("%s: expected status %d, but got %d", test.name, test.expectedStatus, w.Code)
		}
	}
}

func TestDurations(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		min     time.Duration
		max     time.Duration
	}{
		{
			name: "no delay",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("OK"))
			},
			min: time.Duration(0),
			max: time.Millisecond * 100,
		},
		{
			name: "100ms delay",
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(time.Millisecond * 100)
				w.Write([]byte("OK"))
			},
			min: time.Millisecond * 100,
			max: time.Millisecond * 200,
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)

		var actualDuration time.Duration

		h := Handler{
			Next: test.handler,
			Logger: func(url *url.URL, status int, len int64, d time.Duration) {
				actualDuration = d
			},
			Skip: SkipHealthEndpoint,
		}
		h.ServeHTTP(w, r)

		if actualDuration < test.min || actualDuration > test.max {
			t.Errorf("%s: expected duration between %v and %v, but got %v", test.name, test.min, test.max, actualDuration)
		}
	}
}

func TestJSONLogMessage(t *testing.T) {
	tests := []struct {
		name     string
		now      func() time.Time
		url      string
		status   int
		written  int64
		duration time.Duration
		expected string
	}{
		{
			name:     "basic",
			now:      func() time.Time { return time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC) },
			url:      "/test",
			status:   200,
			written:  454,
			duration: time.Millisecond * 300,
			expected: `{ "time": "2000-01-02T03:04:05Z", "src": "rl", "status": 200, "http_2xx": 1, "len": 454, "ms": 300, "path": "/test" }`,
		},
	}

	for _, test := range tests {
		u, err := url.Parse(test.url)
		if err != nil {
			t.Fatalf("%s: failed to parse URL '%v' with error: %v", test.name, test.url, err)
		}

		actual := JSONLogMessage(test.now, u, test.status, test.written, test.duration)
		if test.expected != actual {
			t.Errorf("%s: expected '%v', got: '%v'", test.name, test.expected, actual)
		}
		valid := json.Valid([]byte(actual))
		if !valid {
			t.Errorf("%s: failed to parse JSON message '%v'", test.name, actual)
		}
	}
}
