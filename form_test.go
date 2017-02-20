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
