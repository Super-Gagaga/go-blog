package slug

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode"
)

var hyphenPattern = regexp.MustCompile(`-+`)

// maxSlugRunes 截断上限:需同时满足 articles.slug VARCHAR(300)
// 与 categories.slug VARCHAR(100)(随机后缀最长再占 7 字符)
const maxSlugRunes = 64

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
	if runes := []rune(out); len(runes) > maxSlugRunes {
		out = strings.Trim(string(runes[:maxSlugRunes]), "-")
	}
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
