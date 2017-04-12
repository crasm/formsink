package main

import (
	"io"
	"path/filepath"
	"strings"

	"flag"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/crasm/formsink/lib"
)

var listen = flag.String("listen", "localhost:1234", "Address and port to bind to.")
var maildir = flag.String("maildir", lib.DefaultMaildirPath, "Path to the MAILDIR where incoming messages will be stored; will be created if it does not exist.")
var redirect = flag.String("redirect", "", "URL to redirect the user to after submitting the form. Strongly recommended to be set to a confirmation page to avoid multiple submissions.")

var insecure = flag.Bool("insecure", false, "Use HTTP (insecure) rather than HTTPS.")
var tlsCert = flag.String("tls-cert", "", "Certificate file as documented in https://golang.org/pkg/net/http/#ListenAndServeTLS.")
var tlsKey = flag.String("tls-key", "", "Private key file as document in https://golang.org/pkg/net/http/#ListenAndServeTLS.")

func main() {
	flag.Parse()

	filepaths := []string{}
	for _, arg := range flag.Args() {
		filepath.Walk(arg, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				logrus.WithFields(logrus.Fields{
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
			logrus.WithFields(logrus.Fields{
				"file": file,
			}).Error(err)
			continue
		}
		defer r.Close()

		readers = append(readers, r)
	}

	sink, err := lib.NewSinkFromReader(*maildir, *redirect, readers...)
	if err != nil {
		logrus.Fatal(err)
	}

	if !*insecure {
		okCert := *tlsCert != ""
		okKey := *tlsKey != ""
		if !(okCert && okKey) {
			logrus.WithFields(logrus.Fields{
				"okCert": okCert,
				"okKey":  okKey,
			}).Fatal("Missing configuration for TLS. Please provide a certificate and private key or disable TLS using the --insecure flag.")
		}

		if !strings.HasSuffix(*listen, ":443") {
			logrus.Warn("Not listening on standard HTTPS port 443")
		}

		logrus.WithFields(logrus.Fields{
			"address": *listen,
		}).Info("Listening for HTTPS requests")
		logrus.Fatal(http.ListenAndServeTLS(*listen, *tlsCert, *tlsKey, sink))

	} else {
		if !strings.HasSuffix(*listen, ":80") {
			logrus.Warn("Not listening on standard HTTP port 80")
		}

		logrus.WithFields(logrus.Fields{
			"address": *listen,
		}).Info("Listening for HTTP requests")
		logrus.Fatal(http.ListenAndServe(*listen, sink))
	}

}
