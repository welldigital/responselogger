package responselogger

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
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
			expectedStatus:        50,
			expectedLength:        len(httpResponseBody),
			expectedBody:          httpResponseBody,
			r:                     httptest.NewRequest(http.MethodGet, "/index.html", nil),
			expectedMessageLogged: true,
		},
		{
			name: ">= 100, < 200",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(101)
				w.Write([]byte(httpResponseBody))
			},
			expectedStatus:        101,
			expectedLength:        len(httpResponseBody),
			expectedBody:          httpResponseBody,
			r:                     httptest.NewRequest(http.MethodGet, "/index.html", nil),
			expectedMessageLogged: true,
		},
		{
			name: "200 (default)",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(httpResponseBody))
			},
			expectedStatus:        200,
			expectedLength:        len(httpResponseBody),
			expectedBody:          httpResponseBody,
			r:                     httptest.NewRequest(http.MethodGet, "/index.html", nil),
			expectedMessageLogged: true,
		},
		{
			name: "200 (specific)",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				w.Write([]byte(httpResponseBody))
			},
			expectedStatus:        200,
			expectedLength:        len(httpResponseBody),
			expectedBody:          httpResponseBody,
			r:                     httptest.NewRequest(http.MethodGet, "/index.html", nil),
			expectedMessageLogged: true,
		},
		{
			name: ">= 200, < 300",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(201)
				w.Write([]byte(httpResponseBody))
			},
			expectedStatus:        201,
			expectedLength:        len(httpResponseBody),
			expectedBody:          httpResponseBody,
			r:                     httptest.NewRequest(http.MethodGet, "/index.html", nil),
			expectedMessageLogged: true,
		},
		{
			name: "redirect",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/other", http.StatusMovedPermanently)
			},
			expectedStatus:        301,
			expectedLength:        len("<a href=\"/other\">Moved Permanently</a>.\n\n"),
			expectedBody:          "<a href=\"/other\">Moved Permanently</a>.\n\n",
			r:                     httptest.NewRequest(http.MethodGet, "/index.html", nil),
			expectedMessageLogged: true,
		},
		{
			name: "http 500 error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "error!", 500)
			},
			expectedStatus:        500,
			expectedLength:        len("error!\n"),
			expectedBody:          "error!\n",
			r:                     httptest.NewRequest(http.MethodGet, "/index.html", nil),
			expectedMessageLogged: true,
		},
		{
			name: "skip health check",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(201)
				w.Write([]byte(httpResponseBody))
			},
			expectedStatus:        201,
			expectedBody:          httpResponseBody,
			r:                     httptest.NewRequest(http.MethodGet, "/health", nil),
			expectedMessageLogged: false,
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()

		var loggedMethod string
		var loggedURL string
		var loggedStatus int
		var loggedLength int64
		var messageLogged bool

		h := Handler{
			Next: test.handler,
			Logger: func(r *http.Request, status int, len int64, d time.Duration) {
				loggedMethod = r.Method
				loggedURL = r.URL.String()
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
			if test.r.Method != loggedMethod {
				t.Errorf("%s: expected method '%v' to be logged, but got '%v'", test.name, test.r.Method, loggedMethod)
			}
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

func TestHandlerDurationLogging(t *testing.T) {
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
			Logger: func(r *http.Request, status int, len int64, d time.Duration) {
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

func TestHeadersAreNotLost(t *testing.T) {
	tests := []struct {
		name            string
		handler         http.HandlerFunc
		expectedHeaders http.Header
	}{
		{
			name: "no headers",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("OK"))
			},
			expectedHeaders: http.Header{},
		},
		{
			name: "add X-Powered-By header",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("X-Powered-By", "Go")
				w.Write([]byte("OK"))
			},
			expectedHeaders: http.Header{
				"X-Powered-By": []string{"Go"},
			},
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)

		h := Handler{
			Next:   test.handler,
			Logger: func(r *http.Request, status int, len int64, d time.Duration) {},
			Skip:   SkipHealthEndpoint,
		}
		h.ServeHTTP(w, r)

		for k, v1 := range test.expectedHeaders {
			var v2 []string
			var ok bool
			if v2, ok = w.HeaderMap[k]; !ok {
				t.Errorf("%s: expected written headers to contain key '%v', but it wasn't written", test.name, k)
				continue
			}
			if !reflect.DeepEqual(v1, v2) {
				t.Errorf("%s: expected written header '%v' to equal '%v', but got '%v'", test.name, k, v1, v2)
			}
		}
	}
}

func TestJSONLogMessage(t *testing.T) {
	tests := []struct {
		name     string
		now      func() time.Time
		method   string
		url      string
		status   int
		written  int64
		duration time.Duration
		expected string
		fields   map[string]string
	}{
		{
			name:     "basic",
			now:      func() time.Time { return time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC) },
			method:   "GET",
			url:      "/test",
			status:   200,
			written:  454,
			duration: time.Millisecond * 300,
			expected: `{"time":"2000-01-02T03:04:05Z","src":"rl","status":200,"http_2xx":1,"len":454,"ms":300,"method":"GET","path":"/test"}` + "\n",
		},
		{
			name:     "404",
			now:      func() time.Time { return time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC) },
			method:   "GET",
			url:      "/test",
			status:   404,
			written:  454,
			duration: time.Millisecond * 300,
			expected: `{"time":"2000-01-02T03:04:05Z","src":"rl","status":404,"http_4xx":1,"len":454,"ms":300,"method":"GET","path":"/test"}` + "\n",
		},
		{
			name:     "out of bounds status code",
			now:      func() time.Time { return time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC) },
			method:   "POST",
			url:      "/test",
			status:   999,
			written:  454,
			duration: time.Millisecond * 300,
			expected: `{"time":"2000-01-02T03:04:05Z","src":"rl","status":999,"http_9xx":1,"len":454,"ms":300,"method":"POST","path":"/test"}` + "\n",
		},
		{
			name:     "additional fields",
			now:      func() time.Time { return time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC) },
			method:   "POST",
			url:      "/test",
			status:   222,
			written:  454,
			duration: time.Millisecond * 300,
			fields: map[string]string{
				"field1": "v1",
				"field2": "v2",
			},
			expected: `{"time":"2000-01-02T03:04:05Z","src":"rl","status":222,"http_2xx":1,"len":454,"ms":300,"method":"POST","path":"/test","field1":"v1","field2":"v2"}` + "\n",
		},
	}

	for _, test := range tests {
		u, err := url.Parse(test.url)
		if err != nil {
			t.Fatalf("%s: failed to parse URL '%v' with error: %v", test.name, test.url, err)
		}

		actual := JSONLogMessage(test.now, test.method, u, test.status, test.written, test.duration, test.fields)
		if test.expected != actual {
			t.Errorf("%s: expected '%v', got: '%v'", test.name, test.expected, actual)
		}
		valid := json.Valid([]byte(actual))
		if !valid {
			t.Errorf("%s: failed to parse JSON message '%v'", test.name, actual)
		}
	}
}

func BenchmarkJSONLogMessage(b *testing.B) {
	m := map[string]string{
		"a": "b",
		"c": "d",
	}
	for i := 0; i < b.N; i++ {
		JSONLogMessage(time.Now, "GET", &url.URL{Path: "/index.html"}, http.StatusOK, 1024, time.Millisecond*50, m)
	}
}

func BenchmarkJSONLogMessageWithHeaders(b *testing.B) {
	logger := NewJSONLoggerWithHeaders("a", "b")
	r := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	r.Header.Add("a", "1")
	r.Header.Add("b", "2")
	r.Header.Add("c", "3")
	for i := 0; i < b.N; i++ {
		logger(r, http.StatusOK, 1024, time.Millisecond*50)
	}
}

func BenchmarkHandler(b *testing.B) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	h := NewHandler(next)
	h.Logger = func(r *http.Request, status int, len int64, d time.Duration) {}

	r := httptest.NewRequest(http.MethodGet, "/index.html", nil)
	w := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		h.ServeHTTP(w, r)
	}
}

func TestJSONEscape(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "example.com",
			expected: "example.com",
		},
		{
			input:    "/",
			expected: "/",
		},
		{
			input:    `/test/"q"`,
			expected: `/test/\"q\"`,
		},
		{
			input:    "\n",
			expected: "",
		},
		{
			input:    "\t",
			expected: "",
		},
		{
			input:    "/test/section",
			expected: "/test/section",
		},
		{
			input:    "/test/中文",
			expected: "/test/中文",
		},
		{
			input:    "/±!@£$^&*()_+/section",
			expected: "/±!@£$^&*()_+/section",
		},
		{
			input:    "search/%20%42",
			expected: "search/%20%42",
		},
		{
			input:    "search/\\/test",
			expected: "search/\\\\/test",
		},
	}

	for _, test := range tests {
		actual := jsonEscape(test.input)
		if test.expected != actual {
			t.Errorf("'%v': expected '%s', got '%s'", test.input, test.expected, actual)
		}
	}
}
