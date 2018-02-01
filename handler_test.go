package responselogger

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestThatHTTPStatusesAreLogged(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.HandlerFunc
		expectedStatus int
		expectedLength int64
	}{
		{
			name: ">= 0,< 100",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(50)
				w.Write([]byte("123"))
			},
			expectedStatus: 50,
			expectedLength: 3,
		},
		{
			name: ">= 100, < 200",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(101)
				w.Write([]byte("123"))
			},
			expectedStatus: 101,
			expectedLength: 3,
		},
		{
			name: "200 (default)",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("123"))
			},
			expectedStatus: 200,
			expectedLength: 3,
		},
		{
			name: "200 (specific)",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				w.Write([]byte("123"))
			},
			expectedStatus: 200,
			expectedLength: 3,
		},
		{
			name: ">= 200, < 300",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(201)
				w.Write([]byte("123"))
			},
			expectedStatus: 201,
			expectedLength: 3,
		},
		{
			name: "http 500 error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "error!", 500)
			},
			expectedStatus: 500,
			expectedLength: int64(len("error!\n")),
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)

		var actualURL *url.URL
		var actualStatus int
		var actualLength int64

		h := Handler{
			Next: test.handler,
			Logger: func(url *url.URL, status int, len int64, d time.Duration) {
				actualURL = url
				actualStatus = status
				actualLength = len
			},
		}
		h.ServeHTTP(w, r)

		expectedURL, err := url.Parse("/test")
		if err != nil {
			t.Fatalf("could not parse test URL with err: %v", err)
		}
		if expectedURL.String() != actualURL.String() {
			t.Errorf("%s: expected URL %v, but got %v", test.name, expectedURL, actualURL)
		}

		if test.expectedStatus != actualStatus {
			t.Errorf("%s: expected status %d, but got %d", test.name, test.expectedStatus, actualStatus)
		}
		if test.expectedLength != actualLength {
			t.Errorf("%s: expected length %d, but got %d", test.name, test.expectedLength, actualLength)
		}
	}
}

func TestJSONLogMessage(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		status   int
		written  int64
		duration time.Duration
		expected string
	}{
		{
			name:     "basic",
			url:      "/test",
			status:   200,
			written:  454,
			duration: time.Millisecond * 200,
			expected: `{ "src": "rl", "status": 200, "http_2xx": 1, "len": 454, "ms": 200, "path": "/test" }`,
		},
	}

	for _, test := range tests {
		u, err := url.Parse(test.url)
		if err != nil {
			t.Fatalf("%s: failed to parse URL '%v' with error: %v", test.name, test.url, err)
		}

		actual := JSONLogMessage(u, test.status, test.written, test.duration)
		if test.expected != actual {
			t.Errorf("%s: expected '%v', got: '%v'", test.name, test.expected, actual)
		}
		valid := json.Valid([]byte(actual))
		if !valid {
			t.Errorf("%s: failed to parse JSON message '%v'", test.name, actual)
		}
	}
}
