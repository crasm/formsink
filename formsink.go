package formsink

import (
	"bytes"
	"fmt"
	"net/http"
	"net/mail"
	"os"

	log "github.com/Sirupsen/logrus"
	gm "github.com/jpoehls/gophermail"
	md "github.com/luksen/maildir"
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

type formSink struct {
	depositor depositor
	redirect  string
	forms     map[string]*Form
}

type Form struct {
	Name   string
	Fields []string
	Files  []string
}

type depositor interface {
	Deposit(*gm.Message) error
}

type maildir struct {
	dir string
}

func (m *maildir) Deposit(msg *gm.Message) error {
	dir := md.Dir(m.dir)
	dir.Create() // TODO err?

	delivery, err := dir.NewDelivery()
	if err != nil {
		return err
	}
	defer delivery.Close()

	msgBytes, err := msg.Bytes()
	if err != nil {
		return err
	}

	_, err = delivery.Write(msgBytes)
	return err
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

// TODO: this method is doing way too much. It should be higher level.
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

	// Begin building the message.
	msg := &gm.Message{
		From: formSinkAddress,
		To: []mail.Address{mail.Address{
			// e.g. contact@example.com
			Address: form.Name + "@" + hostname,
		}},
		Subject:     form.Name + " request",
		Attachments: make([]gm.Attachment, 0, 0),
	}

	// Build message body
	body := &bytes.Buffer{}

	for _, id := range form.Fields {
		// TODO: try to decouple this because needing the directly
		// executed anonymous function is probably a code smell.
		func() {
			body.WriteString(id)
			body.WriteString(": ")
			defer body.WriteString("\n")

			v := r.FormValue(id)
			if v == "" {
				log.WithFields(log.Fields{
					"id": id,
				}).Warn("No value for id")
				return
			}

			body.WriteString(v)
		}()
	}

	msg.Body = body.String()

	// Add files as attachments
	for _, id := range form.Files {
		func() {
			file, meta, err := r.FormFile(id)
			if err != nil {
				log.WithFields(log.Fields{
					"id":    id,
					"error": err.Error(),
				}).Warn("Error parsing file")
				return
			}

			msg.Attachments = append(msg.Attachments,
				gm.Attachment{
					Name: meta.Filename,
					Data: file,
				})
		}()
	}

	if err := fs.depositor.Deposit(msg); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Error while building and saving the message")
		writeStatus(w, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", fs.redirect)
	writeStatus(w, http.StatusSeeOther)

	// TODO: Add test form for happy path of execution
}

func writeStatus(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	fmt.Fprintf(w, "%d", status)
}

func e(format string, a ...interface{}) error {
	return fmt.Errorf("formsink: "+format, a...)
}
