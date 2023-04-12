package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

const url = "https://example.com"
const port = 8080

func render(res http.ResponseWriter, Duration string) {
	res.Header().Set("Content-Type", "image/svg+xml")
	fmt.Fprintln(res, `<svg xmlns="http://www.w3.org/2000/svg" version="1.1" viewBox="0 0 164 164">`)
	fmt.Fprintln(res, `<rect x="0" y="0" width="100%" height="100%" fill="#ffe" />`)
	fmt.Fprintln(res, `<text x="20" y="20" fill="#000">`, Duration, `</text>`)
	fmt.Fprintln(res, `</svg>`)
}

func main() {
	benchmark()
	mux := http.NewServeMux()

	log.Println("Listening on http://127.0.0.1:" + strconv.Itoa(port))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), mux))
}
