package main

import (
	"github.com/crasm/formsink/lib"

	"log"
	"net/http"
	"os"
)

func main() {
	contact, err := os.Open("resources/contact.html")
	if err != nil {
		panic(err)
	}

	sink, err := lib.NewSinkFromReader("https://duckduckgo.com/", contact)
	log.Fatal(http.ListenAndServe("localhost:1234", sink))
}
