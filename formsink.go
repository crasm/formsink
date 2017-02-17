package formsink

import (
	"fmt"
	"net/http"
)

type formSink struct {
	redirect string
	forms    map[string]*Form
}

type Form struct {
	Name   string
	Fields []string
	Files  []string
}

func New(redirect string, forms ...*Form) (http.Handler, error) {
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

	return &formSink{redirect, formMap}, nil
}

func (fs *formSink) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeStatus(w, http.StatusMethodNotAllowed)
		return
	}

	_, ok := fs.forms[r.URL.Path[1:]] // Path starts with '/'
	if !ok {
		writeStatus(w, http.StatusNotFound)
		return
	}

	w.Header().Set("Location", fs.redirect)
	writeStatus(w, http.StatusSeeOther)
}

func writeStatus(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	fmt.Fprintf(w, "%d", status)
}

func e(format string, a ...interface{}) error {
	return fmt.Errorf("formsink: "+format, a...)
}
