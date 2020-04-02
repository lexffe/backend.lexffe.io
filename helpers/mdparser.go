package helpers

import (
	"bytes"
	mdlib "github.com/gomarkdown/markdown"
	"github.com/microcosm-cc/bluemonday"
)

// ParseMD is a helper function to parse markdown into html + sanitising.
func ParseMD(markdown string) (string, error) {
	var html bytes.Buffer

	unsafeHTML := mdlib.ToHTML([]byte(markdown), nil, nil)
	_, err := html.Write(bluemonday.UGCPolicy().SanitizeBytes(unsafeHTML))

	if err != nil {
		return "", err
	}

	return html.String(), nil
}
