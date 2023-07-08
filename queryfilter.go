package queryfilter

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

type ChainingStrategyType string

const (
	ChainingStrategyOr  ChainingStrategyType = "OR"
	ChainingStrategyAnd ChainingStrategyType = "AND"
)

type PlaceholderStrategy int

const (
	PlaceholderStrategyQuestionmark PlaceholderStrategy = iota
	PlaceholderStrategyColon
	PlaceholderStrategyDollar
)

var (
	// TagName defines the struct tag we look for in the structs we're parsing
	TagName = "filter"

	// Operators is a globally defined map of available operators
	Operators = map[string]Operator{}

	// DefaultChainingStrategy defined if clauses are glued together using an `OR` statement
	// of an `AND` statement. We default to using `AND` but can be configured to match
	// desired behavior.
	DefaultChainingStrategy = ChainingStrategyAnd

	// DefaultPlaceholderStrategy defined how we define placeholders in the resulting
	// query. Eg: PlaceholderStategyQuestionmark inserts a single questionmark as a placeholder.
	// this is being used in MySQL / MariaDB / SQLite databases. Whereas PlaceholderStrategyDollar will
	// insert a numbered dollarsign ($1, $2, etc) best used with a PostgreSQL database.
	//
	// It defaults to PlaceholderStategyQuestionmark.
	DefaultPlaceholderStrategy = PlaceholderStrategyQuestionmark

	// some of the placeholder stategies require an integer value associated with it.
	// eg: $1, $2, $3 etc. Because we don't want to make assumtions on the rest of your query,
	// you are able to set the placeholder query offset.
	//
	// It defaults to 1.
	DefaultPlaceholderStategyNumberingOffset = 1
)

type Clause struct {
	// Col describes the database column the operation works on.
	Col string

	// Op describes what operation to use on the column
	Op string

	// Val holds the value the operation is performed with
	Val any

	// reflected value of the Val field
	reflectedValue reflect.Value
}

func (c *Clause) AssertTypeOneOf(kinds ...reflect.Kind) error {
	actualKind := c.reflectedValue.Kind()

	// do the actual matching
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

type Opts struct {
	ChainingStrategy    ChainingStrategyType
	PlaceholderStrategy PlaceholderStrategy
	PlaceholderOffset   int
}

func DefaultOpts() *Opts {
	return &Opts{
		ChainingStrategy:    DefaultChainingStrategy,
		PlaceholderStrategy: DefaultPlaceholderStrategy,
		PlaceholderOffset:   DefaultPlaceholderStategyNumberingOffset,
	}
}

type OptFn = func(o *Opts)

func WithChainingStrategy(typ ChainingStrategyType) OptFn {
	return func(o *Opts) {
		o.ChainingStrategy = typ
	}
}

func WithPlaceholderStrategy(strategy PlaceholderStrategy) OptFn {
	return func(o *Opts) {
		o.PlaceholderStrategy = strategy
	}
}

func WithPlaceholderOffset(offset int) OptFn {
	return func(o *Opts) {
		o.PlaceholderOffset = offset
	}
}

// ToSql takes a filter struct and returns a parameterized SQL string
// and its values in order to be applied in a query.
func ToSql(f any, fns ...OptFn) (query string, args []any, err error) {
	opts := DefaultOpts()
	for _, fn := range fns {
		fn(opts)
	}

	clauses, err := buildClauses(f)
	if err != nil {
		return "", nil, err
	}

	sql, args, err := toSql(clauses, opts)
	sql = applyPlaceholders(sql, opts)

	return sql, args, err
}

func applyPlaceholders(q string, opts *Opts) string {
	switch opts.PlaceholderStrategy {
	case PlaceholderStrategyQuestionmark:
		return q

	case PlaceholderStrategyColon:
		return replace(q, opts.PlaceholderOffset, colonReplacer)

	case PlaceholderStrategyDollar:
		return replace(q, opts.PlaceholderOffset, dollarReplacer)
	}

	return ""
}

func toSql(clauses []Clause, opts *Opts) (string, []any, error) {
	var (
		segs []string
		args []any
	)

	for _, c := range clauses {
		// skip nil values
		if c.Val == nil {
			continue
		}

		operator, ok := Operators[c.Op]
		if !ok {
			return "", nil, fmt.Errorf("operator %s is not available", c.Op)
		}

		sql, newArgs, err := operator(c)
		if err != nil {
			return "", nil, err
		}

		segs = append(segs, fmt.Sprintf("%s %s", c.Col, sql))
		args = append(args, newArgs...)
	}

	sep := fmt.Sprintf(" %s ", opts.ChainingStrategy)
	return strings.Join(segs, sep), args, nil
}

func buildClauses(f any) ([]Clause, error) {
	t := reflect.TypeOf(f)
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("unable to build filter: provided value is not a struct")
	}

	v := reflect.ValueOf(f)
	fields := reflect.VisibleFields(t)
	clauses := make([]Clause, len(fields))

	for idx, field := range fields {
		tag, ok := field.Tag.Lookup(TagName)
		if !ok {
			continue
		}

		rawValue := v.FieldByName(field.Name)
		if !rawValue.IsValid() {
			continue
		}

		column, operator, err := parseTag(tag)
		if err != nil {
			return nil, err
		}

		val, err := readValue(rawValue)
		if err != nil {
			return nil, err
		}

		clauses[idx] = Clause{
			Col: column,
			Op:  operator,
			Val: val,

			// store the reflected value for later use
			reflectedValue: rawValue.Elem(),
		}
	}

	return clauses, nil
}

func readValue(v reflect.Value) (any, error) {
	// dereference pointer first if applicable
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	if !v.IsValid() {
		return nil, nil
	}

	// then try to determine the type and return the correct type
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), nil

	case reflect.Float32, reflect.Float64:
		return v.Float(), nil

	case reflect.String:
		return v.String(), nil

	case reflect.Bool:
		return v.Bool(), nil

	case reflect.Array, reflect.Slice:
		args := make([]string, v.Len())
		for i := 0; i < v.Len(); i++ {
			args[i] = v.Index(i).String()
		}
		return args, nil

	case reflect.Struct:
		// not parsing (custom) structs at this time,
		// with the only exception being the time.Time
		if t, ok := v.Interface().(time.Time); ok {
			return t, nil
		}
		return nil, fmt.Errorf("structs are not supported, only time.Time")

	default:
		return nil, fmt.Errorf("unsupported type: %v", v.Kind())
	}
}

func parseTag(tag string) (column, operator string, err error) {
	col, operator, found := strings.Cut(tag, ",")

	// if theres no operator defined, default to equality
	if !found {
		return col, "eq", nil
	}

	// split the string eg: `op=eq`to get the name of the operator here
	_, op, found := strings.Cut(operator, "=")
	if !found {
		return "", "", fmt.Errorf("incorrectly formatted tag: %s", tag)
	}

	return col, strings.TrimSpace(op), nil
}

// readSliceElems takes a reflect.Value of a slice/array
// and returns all elements in that slice/array as a slice.
func readSliceElems(v reflect.Value) ([]any, error) {
	if v.Len() <= 0 {
		return []any{}, nil
	}

	out := make([]any, v.Len())
	for i := 0; i < v.Len(); i++ {
		val, err := readValue(v.Index(i))
		if err != nil {
			return nil, err
		}
		out[i] = val
	}

	return out, nil
}
