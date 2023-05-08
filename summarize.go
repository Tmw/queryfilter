package queryfilter

import (
	"fmt"
	"strings"
)

func summarize[T fmt.Stringer](items ...T) string {
	if len(items) == 1 {
		return items[0].String()
	}

	var b strings.Builder
	for i, k := range items {
		if i > 0 && i < len(items)-1 {
			b.WriteString(", ")
		}

		if i == len(items)-1 {
			b.WriteString(" or ")
		}

		b.WriteString(k.String())
	}

	return b.String()
}
