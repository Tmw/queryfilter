package queryfilter

import (
	"fmt"
	"reflect"
)

// Clause holds the fields that are parsed from the original QueryFilter struct fields.
//
// The only outside use of this type is when defining a custom operator,
// as an operator is defined as a function that takes a Clause and returns the query segment,
// the values to be used in the query, and optionally an error.
type Clause struct {
	// Col describes the database column the operation works on.
	Col string

	// Op describes what operation to use on the column (eg: eq, gt, lt, etc)
	Op string

	// Val holds the value the operation is performed with
	Val any

	// cached reflected value of the Val field
	reflectedValue reflect.Value
}

// AssertTypeOneOf checks if the Clause's reflected value is one of the provided kinds.
//
// This function is used in custom operators to check if the provided field in the QueryFilter struct
// is of a type that the operator can work on. For example for the use of the `in` or `between` operator,
// a slice or array type is expected.
//
// the function returns an error if there's a mismatch in the types.
func (c *Clause) AssertTypeOneOf(kinds ...reflect.Kind) error {
	actualKind := c.reflectedValue.Kind()

	for _, k := range kinds {
		if k == actualKind {
			return nil
		}
	}

	return fmt.Errorf(
		"expected %s; got %s for operation %s",
		summarize(kinds...),
		actualKind,
		c.Op,
	)
}
