package queryfilter

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToSQLSimpleTypes(t *testing.T) {
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

func TestToSQLSimpleTypesNoPointers(t *testing.T) {
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

func TestToSQLWithSlice(t *testing.T) {
	type filter struct {
		Colors []string `filter:"color,op=in"`
		Brands []string `filter:"brand,op=not-in"`
	}

	f := filter{
		Colors: []string{
			"yellow",
			"orange",
			"hot-pink",
		},
		Brands: []string{
			"wrong-brand",
		},
	}

	q, v, e := ToSQL(f)
	eq := `color IN(?,?,?) AND brand NOT IN(?)`
	ev := []any{"yellow", "orange", "hot-pink", "wrong-brand"}

	assert.Nil(t, e)
	assert.EqualValues(t, eq, q)
	assert.ElementsMatch(t, ev, v)
}

func TestToSQLWithEmptySlice(t *testing.T) {
	type filter struct {
		Colors []string `filter:"color,op=in"`
		Brands []string `filter:"brand,op=not-in"`
	}

	f := filter{}

	q, v, e := ToSQL(f)
	eq := `color IN(NULL) AND brand NOT IN(NULL)`
	ev := []any{}

	assert.Nil(t, e)
	assert.EqualValues(t, eq, q)
	assert.Empty(t, v)
	assert.ElementsMatch(t, ev, v)
}

func TestToSQLBetween(t *testing.T) {
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

func TestToSQLBetweenNotEnoughParams(t *testing.T) {
	type filter struct {
		PriceRange *[]float64 `filter:"price,op=between"`
	}
	f2 := filter{
		PriceRange: &[]float64{},
	}

	q, v, e := ToSQL(f2)
	assert.Errorf(t, e, "between expects two elements in its slice")
	assert.Equal(t, "", q)
	assert.ElementsMatch(t, []float64{}, v)
}

func TestToSQLIsNull(t *testing.T) {
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

func TestToSQLIsNotNull(t *testing.T) {
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

func TestToSQLWithDate(t *testing.T) {
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

func TestToSQLInWrongType(t *testing.T) {
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
