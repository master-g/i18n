package model

import "strings"

func EscapeString(raw string) string {
	sb := &strings.Builder{}
	for i, r := range raw {

		skip := false
		switch r {
		case '@', '?', '\'', '"':
			if i == 0 || (i > 0 && raw[i-1] != '\\') {
				sb.WriteRune('\\')
			}
		case '<':
			sb.WriteString("&lt;")
			skip = true
		case '&':
			sb.WriteString("&amp;")
			skip = true
		}

		if !skip {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
