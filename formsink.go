package formsink

import (
	"fmt"
	"net/http"
)

type FormSink struct {
	forms map[string]*Form
}

type Form struct {
	Name   string
	Fields []string
	Files  []string
}

func (fs *FormSink) AddForm(form *Form) error {
	if form == nil {
		return e("form must not be nil")
	} else if form.Name == "" {
		return e("Form.Name must not be \"\"")
	}

	if fs.forms == nil {
		fs.forms = make(map[string]*Form)
	}

	fs.forms[form.Name] = form
	return nil
}

func (fs *FormSink) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeStatus(w, http.StatusMethodNotAllowed)
		return
	}

	_, ok := fs.forms[r.URL.Path[1:]] // Path starts with '/'
	if !ok {
		writeStatus(w, http.StatusNotFound)
		return
	}

	writeStatus(w, http.StatusSeeOther)
}

func writeStatus(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	fmt.Fprintf(w, "%d", status)
}

func e(format string, a ...interface{}) error {
	return fmt.Errorf("formsink: "+format, a...)
}
