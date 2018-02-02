# responselogger

Server middleware to log HTTP server response status codes and their times.

## Usage

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/a-h/responselogger"
)

func main() {
	// Create a mux to store routes.
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world!"))
	})

	mux.HandleFunc("/other", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	// Wrap the mux inside the responselogger.
	loggedHandler := responselogger.NewHandler(mux)

	// Start serving.
	fmt.Println("Listening on :1234...")
	http.ListenAndServe(":1234", loggedHandler)
	fmt.Println("Exited")
}
```

## Output

### Example output from JSON logging

```json
{"time":"2018-02-01T18:41:31Z","src":"rl","status":404,"http_4xx":1,"len":19,"ms":2,"path":"/other"}
{"time":"2018-02-01T18:41:39Z","src":"rl","status":200,"http_2xx":1,"len":12,"ms":4,"path":"/"}
```