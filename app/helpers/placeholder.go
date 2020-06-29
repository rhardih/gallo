package helpers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
)

type Placeholder struct {
	Width  int
	Height int
	Color  string
}

func (p Placeholder) DataURI() (uri string, err error) {
	tmplString := `
  <svg width="{{ .Width }}" height="{{ .Height }}" xmlns="http://www.w3.org/2000/svg" version="1.1">
    <rect x="0" y="0" width="100%" height="100%" fill="{{ .Color }}"></rect>
  </svg>
  `
	tmpl, err := template.New("placeholder").Parse(tmplString)
	if err != nil {
		return
	}

	buf := new(bytes.Buffer)

	err = tmpl.Execute(buf, p)
	if err != nil {
		return
	}

	b64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	uri = fmt.Sprintf("data:image/svg+xml;base64,%s", b64)

	return
}
