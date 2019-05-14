package main

import (
	"fmt"
	"net/http"
	"strings"
)

func handler(w http.ResponseWriter, r *http.Request) {
	var resp []string

	resp = append(resp, "Oh, Hello!")

	for name, headers := range r.Header {
		for _, h := range headers {
			resp = append(resp, fmt.Sprintf("%v: %v", name, h))
		}
	}
	// fmt.Fprintf(w, strings.Join(resp, "\n"))
	w.Write([]byte(strings.Join(resp, "\n")))
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8123", nil)
}
