package queryfilter

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToSqlSimpleTypes(t *testing.T) {
	type filter struct {
		Name   *string `filter:"name,op=eq"`
		MinAge *int    `filter:"age,op=gt"`
	}

	name, minAge := "bobby", int(42)
	f := filter{
		Name:   &name,
		MinAge: &minAge,
	}

	q, v, e := ToSQL(f)
	eq := `name = ? AND age > ?`
	ev := []any{"bobby", int64(minAge)}

	assert.Nil(t, e)
	assert.EqualValues(t, eq, q)
	assert.ElementsMatch(t, ev, v)
}

func ToSQLSimpleTypesNoPointers(t *testing.T) {
	type filter struct {
		Name   string `filter:"name,op=eq"`
		MinAge int    `filter:"age,op=gt"`
	}

	name, minAge := "bobby", int(42)
	f := filter{
		Name:   name,
		MinAge: minAge,
	}

	q, v, e := ToSQL(f)
	eq := `name = ? AND age > ?`
	ev := []any{"bobby", int64(minAge)}

	assert.Nil(t, e)
	assert.EqualValues(t, eq, q)
	assert.ElementsMatch(t, ev, v)
}

func ToSQLWithSlice(t *testing.T) {
	type filter struct {
		Colors *[]string `filter:"color,op=in"`
	}

	f := filter{
		Colors: &[]string{
			"yellow",
			"orange",
			"hot-pink",
		},
	}

	q, v, e := ToSQL(f)
	eq := `color IN(?,?,?)`
	ev := []any{"yellow", "orange", "hot-pink"}

	assert.Nil(t, e)
	assert.EqualValues(t, eq, q)
	assert.Len(t, v, 3)
	assert.ElementsMatch(t, ev, v)
}

func TestToSqlBetween(t *testing.T) {
	type filter struct {
		PriceRange *[]float64 `filter:"price,op=between"`
	}
	f := filter{
		PriceRange: &[]float64{10.21, 30.66},
	}

	q, v, e := ToSQL(f)
	assert.Nil(t, e)
	assert.Equal(t, "price BETWEEN ? AND ?", q)
	assert.ElementsMatch(t, []float64{10.21, 30.66}, v)
}

func TestToSqlIsNull(t *testing.T) {
	type filter struct {
		TitleEmpty *bool `filter:"title,op=is-null"`
	}

	trueVal := true
	falseVal := false

	cases := []struct {
		v *bool
		e string
	}{
		{v: &trueVal, e: "title IS NULL"},
		{v: &falseVal, e: "title IS NOT NULL"},
		{v: nil, e: ""},
	}

	for _, c := range cases {
		f := filter{TitleEmpty: c.v}
		q, v, e := ToSQL(f)

		assert.Nil(t, e)
		assert.Equal(t, c.e, q)
		assert.ElementsMatch(t, []any{}, v)
	}
}

func TestToSqlIsNotNull(t *testing.T) {
	t.Parallel()
	type filter struct {
		TitleEmpty *bool `filter:"title,op=not-null"`
	}

	trueVal := true
	falseVal := false

	cases := []struct {
		v *bool
		e string
	}{
		{v: &trueVal, e: "title IS NOT NULL"},
		{v: &falseVal, e: "title IS NULL"},
		{v: nil, e: ""},
	}

	for _, c := range cases {
		f := filter{TitleEmpty: c.v}
		q, v, e := ToSQL(f)

		assert.Nil(t, e)
		assert.Equal(t, c.e, q)
		assert.ElementsMatch(t, []any{}, v)
	}
}

func TestToSqlWithDate(t *testing.T) {
	type filter struct {
		DueBy *time.Time `filter:"due,op=gt"`
	}

	now := time.Now()
	f := filter{
		DueBy: &now,
	}

	q, v, e := ToSQL(f)
	assert.Nil(t, e)
	assert.Equal(t, "due > ?", q)
	assert.ElementsMatch(t, []any{now}, v)
}

func TestToSqlInWrongType(t *testing.T) {
	type filter struct {
		Tags *string `filter:"title,op=in"`
	}

	tags := "uh-oh"
	f := filter{
		Tags: &tags,
	}

	_, _, e := ToSQL(f)
	assert.NotNil(t, e)
	assert.ErrorContainsf(t, e, "slice or array; got string", "wrong error")
}

func TestAssertTypeOneOf(t *testing.T) {
	cases := []struct {
		value       any
		expected    []reflect.Kind
		shouldError bool
	}{
		{
			value:       "str",
			expected:    []reflect.Kind{reflect.Int},
			shouldError: true,
		},
		{
			value:       12,
			expected:    []reflect.Kind{reflect.String, reflect.Bool},
			shouldError: true,
		},
		{
			value:       12,
			expected:    []reflect.Kind{reflect.String, reflect.Bool, reflect.Slice},
			shouldError: true,
		},
		{
			value:       "str",
			expected:    []reflect.Kind{reflect.String},
			shouldError: false,
		},
		{
			value:       12,
			expected:    []reflect.Kind{reflect.Int},
			shouldError: false,
		},
		{
			value:       true,
			expected:    []reflect.Kind{reflect.Bool, reflect.String},
			shouldError: false,
		},
	}

	for _, tc := range cases {
		clause := Clause{reflectedValue: reflect.ValueOf(tc.value)}
		err := clause.AssertTypeOneOf(tc.expected...)
		if tc.shouldError {
			assert.Error(t, err, "should error")
			continue
		}

		assert.Nil(t, err, "expected no error")
	}
}
