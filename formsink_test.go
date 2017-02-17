package formsink

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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
		err := sink.AddForm(f)
		assert.NotNil(t, err)
	}
}

func TestNotPost(t *testing.T) {
	sink := &FormSink{}
	sink.AddForm(simpleForm)

	r := httptest.NewRequest(http.MethodGet, "/contact", nil)
	w := httptest.NewRecorder()
	sink.ServeHTTP(w, r)

	result := w.Result()
	assert.Equal(t, http.StatusMethodNotAllowed, result.StatusCode)
}

func TestSimpleFormSink(t *testing.T) {
	sink := &FormSink{}
	sink.AddForm(simpleForm)
}
