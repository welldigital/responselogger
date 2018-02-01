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
