package formsink

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/mail"
	"os"

	"github.com/PuerkitoBio/goquery"
	log "github.com/Sirupsen/logrus"
	gm "github.com/jpoehls/gophermail"
)

// This is the same default as in net/http/request.go
const defaultMaxMemory = 32 << 20 // 32MB

var hostname string
var formSinkAddress mail.Address

func init() {
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Warn("Couldn't read hostname")

		hostname = "example.com"
	}

	formSinkAddress = mail.Address{
		Name:    "FormSink",
		Address: "FormSink@" + hostname,
	}
}

type Form struct {
	Name   string
	Fields []string
	Files  []string
}

type formSink struct {
	depositor depositor
	redirect  string
	forms     map[string]*Form
}

func NewSink(redirect string, forms ...*Form) (http.Handler, error) {
	return newSink(&maildir{"./Maildir/"}, redirect, forms...)
}

func newSink(depositor depositor, redirect string, forms ...*Form) (http.Handler, error) {
	if len(forms) < 1 {
		return nil, e("must have at least one form")
	} else if redirect == "" {
		return nil, e("must provide a redirect URL")
	}

	formMap := make(map[string]*Form)
	for _, f := range forms {
		if f == nil {
			return nil, e("forms cannot be nil")
		} else if f.Name == "" {
			return nil, e("Form.Name must not be \"\"")
		}
		formMap[f.Name] = f
	}

	return &formSink{depositor, redirect, formMap}, nil
}

func NewSinkFromDocument(redirect string, documents ...*goquery.Document) (http.Handler, error) {
	return newSinkFromDocument(&maildir{"./Maildir/"}, redirect, documents...)
}

func newSinkFromDocument(depositor depositor, redirect string, documents ...*goquery.Document) (http.Handler, error) {
	forms := make([]*Form, 0)
	for _, doc := range documents {
		doc.Find("form").Each(func(i int, f *goquery.Selection) {
			// TODO NEXT: write tests, then come back here and figure
			// out exactly what needs to happen in order to find the
			// forms
		})
	}

	return newSink(depositor, redirect, forms...)
}

func (fs *formSink) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeStatus(w, http.StatusMethodNotAllowed)
		return
	}

	form, ok := fs.forms[r.URL.Path[1:]] // Path starts with '/'
	if !ok {
		writeStatus(w, http.StatusNotFound)
		return
	}

	if err := r.ParseMultipartForm(defaultMaxMemory); err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Error parsing multipart form")
		return
	}

	msg := buildMessage(form, r.MultipartForm)

	if err := fs.depositor.Deposit(msg); err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Error while building and saving the message")
		writeStatus(w, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", fs.redirect)
	writeStatus(w, http.StatusSeeOther)
}

func buildMessage(formSpec *Form, multipartForm *multipart.Form) *gm.Message {
	// Begin building the message.
	msg := &gm.Message{
		From: formSinkAddress,
		To: []mail.Address{mail.Address{
			// e.g. contact@example.com
			Address: formSpec.Name + "@" + hostname,
		}},
		Subject:     formSpec.Name + " request",
		Attachments: make([]gm.Attachment, 0, 0),
	}

	// Build message body
	body := &bytes.Buffer{}

	for _, id := range formSpec.Fields {
		body.WriteString(id)
		body.WriteString(": ")

		values, ok := multipartForm.Value[id]
		if !ok || len(values) < 1 {
			log.WithFields(log.Fields{
				"id": id,
			}).Warn("No value for id")
			continue
		}

		if len(values) > 1 {
			log.WithFields(log.Fields{
				"id": id,
			}).Warn("Multiple values for a single field, ignoring all but the first")
		}

		body.WriteString(values[0])
		body.WriteString("\n")
	}

	msg.Body = body.String()

	// Add files as attachments
	for _, id := range formSpec.Files {
		metas, ok := multipartForm.File[id]
		if !ok || len(metas) < 1 {
			log.WithFields(log.Fields{
				"id": id,
			}).Warn("No file for id")
			continue
		}

		if len(metas) > 1 {
			log.WithFields(log.Fields{
				"id": id,
			}).Warn("Multiple files for a single field, ignoring all but the first")
		}

		meta := metas[0]

		file, err := meta.Open()
		if err != nil {
			log.WithFields(log.Fields{
				"id":    id,
				"error": err.Error(),
			}).Warn("Error opening file")
			continue
		}

		msg.Attachments = append(msg.Attachments,
			gm.Attachment{
				Name: meta.Filename,
				Data: file,
			})
	}

	return msg
}

func writeStatus(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	fmt.Fprintf(w, "%d", status)
}

func e(format string, a ...interface{}) error {
	return fmt.Errorf("formsink: "+format, a...)
}
