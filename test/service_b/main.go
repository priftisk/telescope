package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("Hello from service b\n"))
	})

	log.Fatal(http.ListenAndServe(":4444", nil))
}
