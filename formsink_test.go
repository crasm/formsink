package formsink

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const location = "https://ddg.gg"

var simpleForm = &Form{
	Name:   "contact",
	Fields: []string{"name", "email", "message"},
	Files:  []string{"picture"},
}

// The "happy" best-case-scenario path.
func TestHappy(t *testing.T) {
	sink, err := NewSink(location, simpleForm)
	assert.Nil(t, err)

	firefoxPost, err := os.Open("resources/post")
	assert.Nil(t, err)
	r, err := http.ReadRequest(bufio.NewReader(firefoxPost))
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	sink.ServeHTTP(w, r)

	result := w.Result()
	assert.Equal(t, http.StatusSeeOther, result.StatusCode)
	assert.Equal(t, location, result.Header.Get("Location"))
}

func TestNotFound(t *testing.T) {
	sink, err := NewSink(location, simpleForm)
	assert.Nil(t, err)

	r := httptest.NewRequest(http.MethodPost, "/hello", nil)
	w := httptest.NewRecorder()
	sink.ServeHTTP(w, r)

	result := w.Result()
	assert.Equal(t, http.StatusNotFound, result.StatusCode)
}

func TestAddFormError(t *testing.T) {
	forms := []*Form{
		nil,
		&Form{Name: ""},
	}

	for _, f := range forms {
		_, err := NewSink(location, f)
		assert.NotNil(t, err)
	}
}

func TestNotPost(t *testing.T) {
	sink, err := NewSink(location, simpleForm)
	assert.Nil(t, err)

	r := httptest.NewRequest(http.MethodGet, "/contact", nil)
	w := httptest.NewRecorder()
	sink.ServeHTTP(w, r)

	result := w.Result()
	assert.Equal(t, http.StatusMethodNotAllowed, result.StatusCode)
}
