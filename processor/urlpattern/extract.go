package urlpattern

import (
	"regexp"
	"strings"
)

var uuid = regexp.MustCompile(`^[a-fA-F0-9]{8}\-[a-fA-F0-9]{4}\-[a-fA-F0-9]{4}\-[a-fA-F0-9]{4}\-[a-fA-F0-9]{12}$`)
var integer = regexp.MustCompile(`^\d+$`)

// Extract paths from a URL.
func Extract(url string) string {
	var op []string
	for _, seg := range strings.Split(url, "/") {
		if integer.MatchString(seg) {
			op = append(op, "{integer}")
			continue
		}
		if uuid.MatchString(seg) {
			op = append(op, "{uuid}")
			continue
		}
		op = append(op, seg)
	}
	return strings.Join(op, "/")
}
