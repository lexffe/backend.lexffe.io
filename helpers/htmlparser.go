package helpers

import (
	"bytes"
	mdlib "github.com/gomarkdown/markdown"
	"github.com/microcosm-cc/bluemonday"
)

// ParseMD is a helper function to parse markdown into html + sanitising.
func ParseMD(markdown string) string {
	var html bytes.Buffer

	unsafeHTML := mdlib.ToHTML([]byte(markdown), nil, nil)
	html.Write(bluemonday.UGCPolicy().SanitizeBytes(unsafeHTML))

	return html.String()
}
