package helpers

import (
	"regexp"
	"strings"
)

// ParseKebab is 
func ParseKebab(s string) string {
	// compile regex, find string, join string, return.
	return strings.Join(regexp.MustCompile(`([\w]+)`).FindAllString(s, -1), "-")
}
