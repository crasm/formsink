package lib

import (
	"bufio"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"os"
	"testing"

	"github.com/jpoehls/gophermail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const location = "https://ddg.gg"

var simpleForm = &Form{
	Name:   "contact",
	Fields: []string{"name", "email", "message"},
	Files:  []string{"picture"},
}

// This is a function because the attachments are read by the tests. You
// can only read in the contents of a file once before needing to rewind
// or reset.
func simpleMessage() *gophermail.Message {
	hostname, err := os.Hostname()
	if err != nil {
		panic("can't initialize: " + err.Error())
	}

	msg := &gophermail.Message{}
	msg.From = mail.Address{
		Name:    "FormSink",
		Address: "FormSink@" + hostname,
	}

	msg.To = []mail.Address{mail.Address{
		Address: simpleForm.Name + "@" + hostname,
	}}

	msg.Subject = simpleForm.Name + " request"

	tiny, err := os.Open("../resources/tiny.ppm")
	if err != nil {
		panic("can't initialize: " + err.Error())
	}

	picture := gophermail.Attachment{
		Name:        "tiny.ppm",
		ContentType: "image/x-portable-pixmap",
		Data:        tiny,
	}

	msg.Attachments = []gophermail.Attachment{picture}
	msg.Body = "name: crasm\nemail: crasm@formsink.email.vczf.io\nmessage: I &#9829; formsink!\n"
	return msg
}

type mockDepositor struct {
	msg *gophermail.Message
}

func (m *mockDepositor) Deposit(msg *gophermail.Message) error {
	m.msg = msg
	return nil
}

// Replays the captured HTTP POST request against the http.Handler and
// returns the response.
func post(t *testing.T, sink http.Handler) *http.Response {
	firefoxPost, err := os.Open("../resources/post")
	require.Nil(t, err)
	r, err := http.ReadRequest(bufio.NewReader(firefoxPost))
	require.Nil(t, err)

	w := httptest.NewRecorder()
	sink.ServeHTTP(w, r)

	return w.Result()
}

// Checks a message for equality with simpleMessage.
func checkMessage(t *testing.T, msg *gophermail.Message) {
	simpleMessage := simpleMessage()

	require.NotNil(t, msg, "No mail message was provided")
	assert.Equal(t, simpleMessage.From, msg.From)
	assert.Equal(t, simpleMessage.To, msg.To)
	assert.Equal(t, simpleMessage.Subject, msg.Subject)
	assert.Equal(t, simpleMessage.Body, msg.Body)

	assert.Equal(t, len(simpleMessage.Attachments), len(msg.Attachments),
		"Does have the correct number of attachments")
	for i := range msg.Attachments {
		simpleAttachment := simpleMessage.Attachments[i]
		mockAttachment := msg.Attachments[i]

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

	result := post(t, sink)
	assert.Equal(t, http.StatusSeeOther, result.StatusCode)
	assert.Equal(t, location, result.Header.Get("Location"))
	checkMessage(t, mockDepositor.msg)
}

func TestNotFound(t *testing.T) {
	sink, err := NewSink(DefaultMaildirPath, location, simpleForm)
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
		_, err := NewSink(DefaultMaildirPath, location, f)
		assert.NotNil(t, err)
	}
}

func TestNotPost(t *testing.T) {
	sink, err := NewSink(DefaultMaildirPath, location, simpleForm)
	assert.Nil(t, err)

	r := httptest.NewRequest(http.MethodGet, "/contact", nil)
	w := httptest.NewRecorder()
	sink.ServeHTTP(w, r)

	result := w.Result()
	assert.Equal(t, http.StatusMethodNotAllowed, result.StatusCode)
}

func TestNoRedirect(t *testing.T) {
	mockDepositor := &mockDepositor{}

	sink, err := newSink(mockDepositor, location, simpleForm)
	require.Nil(t, err)

	result := post(t, sink)
	assert.Equal(t, http.StatusSeeOther, result.StatusCode)
	assert.Equal(t, location, result.Header.Get("Location"))
	checkMessage(t, mockDepositor.msg)
}
