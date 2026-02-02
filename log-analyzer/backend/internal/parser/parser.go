package parser

import (
	"regexp"
	"strings"
)

func ParseLogLineToSummary(line string) string {
	if line == "" {
		return ""
	}
	s := strings.TrimSpace(line)
	if idx := strings.Index(s, "]: "); idx >= 0 {
		s = strings.TrimSpace(s[idx+3:])
	} else if idx := strings.Index(s, ": "); idx >= 0 {
		after := strings.TrimSpace(s[idx+2:])
		if after != "" {
			s = after
		}
	}
	re := regexp.MustCompile(`^[\w\-\.]+\[\d+\]:?\s*`)
	s = re.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	re = regexp.MustCompile(`\s+`)
	s = re.ReplaceAllString(s, " ")

	if s == "" {
		return line
	}
	return s
}
