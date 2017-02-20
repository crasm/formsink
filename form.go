package formsink

import (
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

type Form struct {
	Name   string
	Fields []string
	Files  []string
}

func documentsToForms(documents ...*goquery.Document) ([]*Form, error) {
	forms := make([]*Form, 0)
	for _, doc := range documents {
		var err error

		doc.Find("form").EachWithBreak(func(_ int, sel *goquery.Selection) bool {

			// e.g. action='https://www.example.com/contact'
			action, ok := sel.Attr("action")
			if !ok {
				err = e("No 'action' attribute available for %v", sel)
				return false
			}

			url, err := url.Parse(action)
			if err != nil {
				return false
			}

			f := &Form{
				Name:   url.Path[1:], // e.g. string("/contact")[1:] => "contact"
				Fields: []string{},
				Files:  []string{},
			}
			forms = append(forms, f)

			return true
		})

		if err != nil {
			return nil, err
		}
	}

	return forms, nil
}
