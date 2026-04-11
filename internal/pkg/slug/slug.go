package slug

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode"
)

var hyphenPattern = regexp.MustCompile(`-+`)

func Make(input string) string {
	input = strings.ToLower(strings.TrimSpace(input))
	var b strings.Builder
	lastHyphen := false

	for _, r := range input {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastHyphen = false
		case unicode.IsLetter(r) || unicode.IsNumber(r):
			b.WriteRune(r)
			lastHyphen = false
		default:
			if !lastHyphen {
				b.WriteByte('-')
				lastHyphen = true
			}
		}
	}

	out := strings.Trim(b.String(), "-")
	out = hyphenPattern.ReplaceAllString(out, "-")
	if out == "" {
		out = "article"
	}
	return out
}

func WithRandomSuffix(input string) string {
	suffix := make([]byte, 3)
	if _, err := rand.Read(suffix); err != nil {
		return Make(input)
	}
	return Make(input) + "-" + hex.EncodeToString(suffix)
}
