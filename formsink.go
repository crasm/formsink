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
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405"))
		return
	}
}

func e(format string, a ...interface{}) error {
	return fmt.Errorf("formsink: "+format, a...)
}
