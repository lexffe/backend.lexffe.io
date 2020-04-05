package helpers

import (
	"regexp"
	"strings"
)

// ParseKebab is
func ParseKebab(s string) (string, error) {
	// compile regex, find string, join string, return.

	inst, err := regexp.Compile(`([\w]+)`)

	if err != nil {
		return "", err
	}

	return strings.ToLower(strings.Join(inst.FindAllString(s, -1), "-")), nil
}
