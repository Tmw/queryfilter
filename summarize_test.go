package queryfilter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Fruit string

func (f Fruit) String() string {
	return string(f)
}

func TestSummarize(t *testing.T) {
	cases := []struct {
		i []Fruit
		e string
	}{
		{
			i: []Fruit{"apple", "banana", "melon"},
			e: "apple, banana or melon",
		},
		{
			i: []Fruit{"apple", "banana"},
			e: "apple or banana",
		},
		{
			i: []Fruit{"apple"},
			e: "apple",
		},
		{
			i: []Fruit{},
			e: "",
		},
	}

	for _, tc := range cases {
		actual := summarize(tc.i...)
		assert.Equal(t, tc.e, actual)
	}
}
