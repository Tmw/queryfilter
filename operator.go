package queryfilter

import (
	"fmt"
	"reflect"
)

// Operator is a function that receives a clause and returns the query segment
// as a string and a slice of values.
//
// Custom operators can be defined by assigning them by name to the global
// Operators map, eg:
//
// queryfilter.Operators["my-operator"] = func(c Clause) (string, []any, error) {...}
// which can then be used in a filter struct:
//
//	type filter struct {
//		Age *int `filter:"age,op=my-operator"`
//	}
type Operator func(c Clause) (string, []any, error)

func init() {
	Operators = map[string]Operator{
		"eq":  SimpleOperator("= ?"),
		"gt":  SimpleOperator("> ?"),
		"gte": SimpleOperator(">= ?"),
		"lte": SimpleOperator("<= ?"),
		"lt":  SimpleOperator("< ?"),
		"in": func(c Clause) (string, []any, error) {
			if err := c.AssertTypeOneOf(reflect.Slice, reflect.Array); err != nil {
				return "", nil, err
			}

			placeholders := PlaceholderList(c.reflectedValue.Len())
			elems, err := readSliceElems(c.reflectedValue)
			if err != nil {
				return "", nil, err
			}
			return fmt.Sprintf("IN(%s)", placeholders), elems, nil
		},
		"not-in": func(c Clause) (string, []any, error) {
			if err := c.AssertTypeOneOf(reflect.Slice, reflect.Array); err != nil {
				return "", nil, err
			}

			placeholders := PlaceholderList(c.reflectedValue.Len())
			elems, err := readSliceElems(c.reflectedValue)
			if err != nil {
				return "", nil, err
			}
			return fmt.Sprintf("NOT IN(%s)", placeholders), elems, nil
		},
		"between": func(c Clause) (string, []any, error) {
			if err := c.AssertTypeOneOf(reflect.Slice, reflect.Array); err != nil {
				return "", nil, err
			}

			elems, err := readSliceElems(c.reflectedValue)
			if err != nil {
				return "", nil, err
			}

			return "BETWEEN ? AND ?", elems[:2], nil
		},
		"is-null": func(c Clause) (string, []any, error) {
			if c.reflectedValue.Bool() {
				return "IS NULL", []any{}, nil
			}

			return "IS NOT NULL", []any{}, nil
		},
		"not-null": func(c Clause) (string, []any, error) {
			if c.reflectedValue.Bool() {
				return "IS NOT NULL", []any{}, nil
			}

			return "IS NULL", []any{}, nil
		},
	}
}

// SimpleOperator is a shorthand function for creating operators with a one-to-one matching
// between column and value. Examples of these are eq, gt, gte without any custom logic.
//
// eg: SimpleOperator("> ?") will return a function that will return the query segment "> ?"
// and the value of the Clause struct as the argument.
func SimpleOperator(r string) Operator {
	return func(c Clause) (string, []any, error) {
		return r, []any{c.Val}, nil
	}
}
