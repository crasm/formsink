package main

import (
	"io"

	"flag"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/crasm/formsink/lib"
)

var listen = flag.String("listen", "localhost:1234", "Address and port to bind to.")
var maildir = flag.String("maildir", lib.DefaultMaildirPath, "Path to the MAILDIR where incoming messages will be stored; will be created if it does not exist.")
var redirect = flag.String("redirect", "", "URL to redirect the user to after submitting the form. Strongly recommended to be set to a confirmation page to avoid multiple submissions.")

func main() {
	flag.Parse()
	readers := make([]io.Reader, flag.NArg())
	for i := range flag.Args() {
		r, err := os.Open(flag.Arg(i))
		if err != nil {
			log.Fatal(err)
		}
		defer r.Close()
		readers[i] = r
	}

	sink, err := lib.NewSinkFromReader(*maildir, *redirect, readers...)
	if err != nil {
		log.Fatal(err)
	}

	log.WithFields(log.Fields{
		"address": *listen,
	}).Info("Listening")
	log.Fatal(http.ListenAndServe(*listen, sink))
}
