package queryfilter

import (
	"fmt"
	"strings"
)

// PlaceholderList generates a list of n placeholder symbols (?) as a comma separated string.
// eg: PlaceholderList(3) => "?,?,?".
//
// note that these placeholders are internal only and will be replaced by the placeholders
// configured by the PlaceholderStrategy when calling ToSQL.
func PlaceholderList(n int) string {
	if n == 0 || n == 1 {
		return "?"
	}

	return strings.Repeat(",?", n)[1:]
}

type replacerFn = func(int) string

func makeReplacer(prefix string) replacerFn {
	return func(i int) string {
		return fmt.Sprintf("%s%d", prefix, i)
	}
}

var (
	defaultReplacer = func(_ int) string { return "?" }
	dollarReplacer  = makeReplacer("$")
	colonReplacer   = makeReplacer(":")
)

func replace(q string, placeholderNumberOffset int, fn replacerFn) string {
	var (
		readerOffset = 0
		b            strings.Builder
		n            = placeholderNumberOffset
	)

	for {
		// look for first ? in offsetted string
		idx := strings.Index(q[readerOffset:], "?")
		if idx < 0 {
			// write the rest of the original query before breaking
			b.WriteString(q[readerOffset:])
			break
		}

		// grab the chunk from last offset to new ?
		chunk := q[readerOffset:][:idx]

		// increment offset and write new placeholder
		readerOffset += idx + 1
		b.WriteString(chunk)
		b.WriteString(fn(n))

		n++
	}

	return b.String()
}
