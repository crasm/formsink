package formsink

import (
	"os"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocumentsToFormsContact(t *testing.T) {
	html, err := os.Open("resources/contact.html")
	require.Nil(t, err)
	defer html.Close()

	doc, err := goquery.NewDocumentFromReader(html)
	require.Nil(t, err)

	forms, err := documentsToForms(doc)
	require.Nil(t, err)

	assert.Len(t, forms, 1)
	assert.Equal(t, *simpleForm, *forms[0])
}

func TestDocumentsToFormsNoAction(t *testing.T) {
	html, err := os.Open("resources/contact.html")
	require.Nil(t, err)
	defer html.Close()

	doc, err := goquery.NewDocumentFromReader(html)
	require.Nil(t, err)

	functions := []func(*goquery.Selection){
		func(s *goquery.Selection) { s.RemoveAttr("action") },
		func(s *goquery.Selection) { s.SetAttr("action", "") },
		func(s *goquery.Selection) { s.SetAttr("action", ":{D") },
		// No URL.Path component
		func(s *goquery.Selection) { s.SetAttr("action", "localhost") },
	}

	form := doc.Find("form")
	for _, f := range functions {
		f(form)
		_, err = documentsToForms(doc)
		assert.NotNil(t, err)
	}
}
