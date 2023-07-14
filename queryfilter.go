package queryfilter

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// ChainingStrategy defines how clauses are glued together in the resulting querystring.
// Eg: using an `OR` statement or an `AND` statement.
// possible options are ChainStrategyOr and ChainStrategyAnd.
type ChainingStrategy string

const (
	ChainingStrategyOr  ChainingStrategy = "OR"
	ChainingStrategyAnd ChainingStrategy = "AND"
)

// PlaceholderStrategy defines how placeholders are defined in the resulting querystring.
// for example working with MySQL based databases you would want to use PlaceholderStrategyQuestionmark
// as the resulting querystring would include a single questionmark as a placeholder
type PlaceholderStrategy int

const (
	// PlaceholderStrategyQuestionmark will insert a single questionmark as a placeholder.
	// this option is best suited for MySQL / MariaDB / SQLite databases.
	PlaceholderStrategyQuestionmark PlaceholderStrategy = iota

	// PlaceholderStrategyColon will insert a positional placeholder using a colon (:1, :2, etc)
	PlaceholderStrategyColon

	// PlaceholderStrategyDollar will insert a positional placeholder using a dollar sign ($1, $2, etc).
	// Most commonly used with PostgreSQL databases.
	PlaceholderStrategyDollar
)

var (
	// TagName defines the struct tag we look for in the structs we're parsing,
	// eg: the `filter` in `filter:"name,op=eq"`. It can be configured by setting
	// `queryFilter.TagName`, eg: `queryFilter.TagName = "qf"` to the value you desire.
	TagName = "filter"

	// Operators is a globally defined map of available operators.
	// See the Operator type for more info.
	Operators = map[string]Operator{}

	// DefaultChainingStrategy defines how clauses are glued together. Eg: using an `OR` statement
	// or an `AND` statement. Default to using `AND` but can be configured either globally or
	// on an individual basis when calling `ToSQL`.
	DefaultChainingStrategy = ChainingStrategyAnd

	// DefaultPlaceholderStrategy configures how we define placeholders in the resulting query.
	//
	// This can be configured globally or on an individual basis when calling `ToSQL`.
	// It defaults to PlaceholderStategyQuestionmark but is configurable either globally or
	// on an individual basis when calling `ToSQL`.
	DefaultPlaceholderStrategy = PlaceholderStrategyQuestionmark

	// Some of the placeholder stategies are indexed (eg: $1, $2, $3, etc..)
	// To avoid clashes with the rest of your query, the starting index is configurable using this option.
	//
	// It defaults to 1 but is configurable either globally or on an individual basis when calling `ToSQL`.
	DefaultPlaceholderStategyIndexOffset = 1
)

// Opts defines the options that are used when running `ToSQL`.
// the opts are constructed everytime `ToSQL` is called and can be configured through the defaults
// defined globally on this module or overwritten on a case-by-case basis when calling `ToSQL` through the
// use of the `OptFn` type.
//
// eg: to set the chaining strategy to `OR` for this call only:
//
//	_, _, _ := ToSQL(filter, WithChainingStrategy(ChainingStrategyOr))
type Opts struct {
	ChainingStrategy    ChainingStrategy
	PlaceholderStrategy PlaceholderStrategy
	PlaceholderOffset   int
}

func DefaultOpts() *Opts {
	return &Opts{
		ChainingStrategy:    DefaultChainingStrategy,
		PlaceholderStrategy: DefaultPlaceholderStrategy,
		PlaceholderOffset:   DefaultPlaceholderStategyIndexOffset,
	}
}

type OptFn = func(o *Opts)

func WithChainingStrategy(typ ChainingStrategy) OptFn {
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

// ToSQL takes a filter struct and returns a parameterized SQL string
// and its values in order to be applied in a query.
func ToSQL(f any, fns ...OptFn) (query string, args []any, err error) {
	opts := DefaultOpts()
	for _, fn := range fns {
		fn(opts)
	}

	clauses, err := buildClauses(f)
	if err != nil {
		return "", nil, err
	}

	sql, args, err := toSQL(clauses, opts)
	sql = applyPlaceholders(sql, opts)

	return sql, args, err
}

func toSQL(clauses []Clause, opts *Opts) (string, []any, error) {
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

func applyPlaceholders(q string, opts *Opts) string {
	switch opts.PlaceholderStrategy {
	case PlaceholderStrategyQuestionmark:
		return replace(q, opts.PlaceholderOffset, defaultReplacer)

	case PlaceholderStrategyColon:
		return replace(q, opts.PlaceholderOffset, colonReplacer)

	case PlaceholderStrategyDollar:
		return replace(q, opts.PlaceholderOffset, dollarReplacer)
	}

	return ""
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

			// store the dereferenced reflected value for later use
			reflectedValue: derefIfApplicable(rawValue),
		}
	}

	return clauses, nil
}

func derefIfApplicable(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		return v.Elem()
	}
	return v
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

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint(), nil

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
