package queryfilter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplace_Default(t *testing.T) {
	query := "name = ? AND color = ?"
	q := replace(query, 1, dollarReplacer)
	e := "name = $1 AND color = $2"

	assert.Equal(t, e, q)
}

func TestReplace_InOperator(t *testing.T) {
	query := "color IN(?, ?, ?)"
	q := replace(query, 1, dollarReplacer)
	e := "color IN($1, $2, $3)"

	assert.Equal(t, e, q)
}

func TestReplace_WithOffset(t *testing.T) {
	query := "name = ? AND color = ?"
	q := replace(query, 5, dollarReplacer)
	e := "name = $5 AND color = $6"

	assert.Equal(t, e, q)
}

func TestReplace_DefaultReplacer(t *testing.T) {
	query := "name = ? AND color = ?"
	q := replace(query, 0, defaultReplacer)
	e := "name = ? AND color = ?"

	assert.Equal(t, e, q)
}

func TestReplace_ColonReplacer(t *testing.T) {
	query := "name = ? AND color = ?"
	q := replace(query, 1, colonReplacer)
	e := "name = :1 AND color = :2"

	assert.Equal(t, e, q)
}

func TestPlaceholderList(t *testing.T) {
	table := []struct {
		num    int
		expect string
	}{
		{0, "?"},
		{1, "?"},
		{2, "?,?"},
		{3, "?,?,?"},
		{4, "?,?,?,?"},
	}

	for _, tc := range table {
		assert.Equal(t, tc.expect, PlaceholderList(tc.num))
	}
}
