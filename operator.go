package queryfilter

import (
	"fmt"
	"reflect"
)

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

func SimpleOperator(r string) Operator {
	return func(c Clause) (string, []any, error) {
		return r, []any{c.Val}, nil
	}
}
