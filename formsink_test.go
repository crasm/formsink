package formsink

import (
	"bufio"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"os"
	"testing"

	"github.com/PuerkitoBio/goquery"
	gm "github.com/jpoehls/gophermail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const location = "https://ddg.gg"

var simpleForm = &Form{
	Name:   "contact",
	Fields: []string{"name", "email", "message"},
	Files:  []string{"picture"},
}

var simpleMessage *gm.Message

func init() {
	msg := &gm.Message{}

	hostname, err := os.Hostname()
	if err != nil {
		panic("can't initialize: " + err.Error())
	}

	msg.From = mail.Address{
		Name:    "FormSink",
		Address: "FormSink@" + hostname,
	}

	msg.To = []mail.Address{mail.Address{
		Address: simpleForm.Name + "@" + hostname,
	}}

	msg.Subject = simpleForm.Name + " request"

	tiny, err := os.Open("resources/tiny.ppm")
	if err != nil {
		panic("can't initialize: " + err.Error())
	}

	picture := gm.Attachment{
		Name:        "tiny.ppm",
		ContentType: "image/x-portable-pixmap",
		Data:        tiny,
	}

	msg.Attachments = []gm.Attachment{picture}

	msg.Body = "name: crasm\nemail: crasm@formsink.email.vczf.io\nmessage: I &#9829; formsink!\n"

	simpleMessage = msg
}

type mockDepositor struct {
	msg *gm.Message
}

func (m *mockDepositor) Deposit(msg *gm.Message) error {
	m.msg = msg
	return nil
}

func checkContactForm(t *testing.T, mockDepositor *mockDepositor, sink http.Handler) {
	firefoxPost, err := os.Open("resources/post")
	require.Nil(t, err)
	r, err := http.ReadRequest(bufio.NewReader(firefoxPost))
	require.Nil(t, err)

	w := httptest.NewRecorder()
	sink.ServeHTTP(w, r)

	result := w.Result()
	assert.Equal(t, http.StatusSeeOther, result.StatusCode)
	assert.Equal(t, location, result.Header.Get("Location"))

	require.NotNil(t, mockDepositor.msg, "No mail message was provided")
	assert.Equal(t, simpleMessage.From, mockDepositor.msg.From)
	assert.Equal(t, simpleMessage.To, mockDepositor.msg.To)
	assert.Equal(t, simpleMessage.Subject, mockDepositor.msg.Subject)
	assert.Equal(t, simpleMessage.Body, mockDepositor.msg.Body)

	assert.Equal(t, len(simpleMessage.Attachments), len(mockDepositor.msg.Attachments))
	for i := range mockDepositor.msg.Attachments {
		simpleAttachment := simpleMessage.Attachments[i]
		mockAttachment := mockDepositor.msg.Attachments[i]

		assert.Equal(t, simpleAttachment.Name, mockAttachment.Name)
		simpleData, err := ioutil.ReadAll(simpleAttachment.Data)
		assert.Nil(t, err)
		mockData, err := ioutil.ReadAll(mockAttachment.Data)
		assert.Nil(t, err)

		assert.Equal(t, simpleData, mockData)
	}
}

// The "happy" best-case-scenario path.
func TestHappy(t *testing.T) {
	mockDepositor := &mockDepositor{}

	sink, err := newSink(mockDepositor, location, simpleForm)
	require.Nil(t, err)

	checkContactForm(t, mockDepositor, sink)
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

func TestDocument(t *testing.T) {
	mockDepositor := &mockDepositor{}

	html, err := os.Open("resources/contact.html")
	require.Nil(t, err)
	doc, err := goquery.NewDocumentFromReader(html)
	require.Nil(t, err)

	sink, err := newSinkFromDocument(mockDepositor, location, doc)
	require.Nil(t, err)

	checkContactForm(t, mockDepositor, sink)
}
