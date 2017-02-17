package formsink

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var simpleForm = &Form{
	Name:   "contact",
	Fields: []string{"name", "email", "message"},
	Files:  []string{"picture"},
}

func TestAddFormError(t *testing.T) {
	forms := []*Form{
		nil,
		&Form{Name: ""},
	}

	for _, f := range forms {
		sink := &FormSink{}
		if err := sink.AddForm(f); err == nil {
			t.Error("Expected an error")
		}
	}
}

func TestNotPost(t *testing.T) {
	sink := &FormSink{}
	sink.AddForm(simpleForm)

	r := httptest.NewRequest(http.MethodGet, "/contact", nil)
	w := httptest.NewRecorder()
	sink.ServeHTTP(w, r)

	result := w.Result()
	if result.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Got %v, expected %v",
			result.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestSimpleFormSink(t *testing.T) {
	sink := &FormSink{}
	sink.AddForm(simpleForm)
}
