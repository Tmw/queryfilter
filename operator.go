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

// RegisterOperator registers an operator with the given name and function.
// the name, given to the operator here, can be used to reference the operator
// from the struct tag.

// Note that calling this function multiple times with the same name will
// overwrite the function previously registered to the operator without warning.
//
// Example:
//
//	RegisterOperator("eq", func()...)
//
//	type Filter struct {
//	    Price int `filter:"price,op=eq"`
//	                                ^^--- operator name
//	}
func RegisterOperator(name string, op Operator) {
	Operators[name] = op
}

func init() {
	// register built in operators
	RegisterOperator("eq", SimpleOperator("= ?"))
	RegisterOperator("gt", SimpleOperator("> ?"))
	RegisterOperator("gte", SimpleOperator(">= ?"))
	RegisterOperator("lte", SimpleOperator("<= ?"))
	RegisterOperator("lt", SimpleOperator("< ?"))

	RegisterOperator("in", func(c Clause) (string, []any, error) {
		if err := c.AssertTypeOneOf(reflect.Slice, reflect.Array); err != nil {
			return "", nil, err
		}

		// early return when passed slice is empty
		if c.reflectedValue.Len() == 0 {
			return "IN(NULL)", []any{}, nil
		}

		placeholders := PlaceholderList(c.reflectedValue.Len())
		elems, err := readSliceElems(c.reflectedValue)

		if err != nil {
			return "", nil, err
		}
		return fmt.Sprintf("IN(%s)", placeholders), elems, nil
	})

	RegisterOperator("not-in", func(c Clause) (string, []any, error) {
		if err := c.AssertTypeOneOf(reflect.Slice, reflect.Array); err != nil {
			return "", nil, err
		}

		// early return when passed slice is empty
		if c.reflectedValue.Len() == 0 {
			return "NOT IN(NULL)", []any{}, nil
		}

		placeholders := PlaceholderList(c.reflectedValue.Len())
		elems, err := readSliceElems(c.reflectedValue)
		if err != nil {
			return "", nil, err
		}
		return fmt.Sprintf("NOT IN(%s)", placeholders), elems, nil
	})

	RegisterOperator("between", func(c Clause) (string, []any, error) {
		if err := c.AssertTypeOneOf(reflect.Slice, reflect.Array); err != nil {
			return "", nil, err
		}

		if c.reflectedValue.Len() < 2 {
			return "", nil, fmt.Errorf("operation between expects two elements in its slice")
		}

		elems, err := readSliceElems(c.reflectedValue)
		if err != nil {
			return "", nil, err
		}

		return "BETWEEN ? AND ?", elems[:2], nil
	})

	RegisterOperator("is-null", func(c Clause) (string, []any, error) {
		if c.reflectedValue.Bool() {
			return "IS NULL", []any{}, nil
		}

		return "IS NOT NULL", []any{}, nil
	})

	RegisterOperator("not-null", func(c Clause) (string, []any, error) {
		if c.reflectedValue.Bool() {
			return "IS NOT NULL", []any{}, nil
		}

		return "IS NULL", []any{}, nil
	})
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
