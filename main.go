package main

import (
	"io"
	"path/filepath"

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

	filepaths := []string{}
	for _, arg := range flag.Args() {
		filepath.Walk(arg, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.WithFields(log.Fields{
					"path": path,
				}).Error(err)
				return err
			}

			if info.IsDir() {
				return nil // Skip directories because we only care about html files.
			}

			filepaths = append(filepaths, path)
			return nil
		})
	}

	readers := []io.Reader{}
	for _, file := range filepaths {
		r, err := os.Open(file)
		if err != nil {
			log.WithFields(log.Fields{
				"file": file,
			}).Error(err)
			continue
		}
		defer r.Close()

		readers = append(readers, r)
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
