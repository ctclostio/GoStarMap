package celestial

import (
	"fmt"
	"strings"
)

// FormatNumber adds commas to numbers for readability. Handles negatives and
// trims trailing zeros (and a trailing decimal point) after %.2f formatting.
func FormatNumber(num float64) string {
	str := fmt.Sprintf("%.2f", num)
	if strings.Contains(str, ".") {
		str = strings.TrimRight(strings.TrimRight(str, "0"), ".")
	}

	prefix := ""
	if strings.HasPrefix(str, "-") {
		prefix = "-"
		str = str[1:]
	}

	parts := strings.SplitN(str, ".", 2)
	intPart := parts[0]
	decPart := ""
	if len(parts) > 1 {
		decPart = "." + parts[1]
	}

	n := len(intPart)
	if n <= 3 {
		return prefix + intPart + decPart
	}

	var b strings.Builder
	b.Grow(n + n/3 + 2)
	b.WriteString(prefix)
	for i, c := range intPart {
		if i > 0 && (n-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	b.WriteString(decPart)
	return b.String()
}
