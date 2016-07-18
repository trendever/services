package utils

import (
	"strings"
)

// SplitAndTrim splits string by "," and trims spaces; slice is returned
//  empty elements are omitted
func SplitAndTrim(to string) (out []string) {
	for _, reciever := range strings.Split(to, ",") {
		if reciever == "" {
			continue
		}
		out = append(out, strings.TrimSpace(reciever))
	}

	return
}
