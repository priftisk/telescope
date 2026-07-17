package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Serving", r.URL.Path)
		w.WriteHeader(200)
		w.Write([]byte("Hello from service a\n"))
	})
	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("Hello from /users endpoint."))
	})

	log.Fatal(http.ListenAndServe(":3333", nil))
}
