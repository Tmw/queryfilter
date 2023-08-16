# Queryfilter

Construct parameterized SQL `where` clauses using Go structs.

## Installing

```golang
import "github.com/tmw/queryfilter"
```

```console
go get github.com/tmw/queryfilter
```

## Usage

```golang
// define a struct with fields to filter on
type Filter struct {
	Sizes 	 []string `filter:"size,op=in"`
	PriceMin int 	  `filter:"price,op=gte"`
	PriceMax int      `filter:"price,op=lte"`
}

// pass the right values
f := Filter{
	Sizes: []string{"L", "XL"},
	PriceMin: 15,
	PriceMax: 35,
}

// turn into parameterized SQL statement
query, params, err := queryfilter.ToSQL(f)
if err != nil {
	log.Fatal(fmt.Errorf("error building query: %w", err)))
}

// Results in:
// query = sizes IN(?, ?) AND price > ? AND price < ?
// params = []any{"L", "XL", 15, 35}

// passing it to the DB layer:
query = fmt.Sprintf("SELECT * FROM tshirts WHERE %s", query)
rows, err := db.Query(query, params...)
```

## Running examples

```console
# to run the sqlite example:
make example-sqlite

# to run the mysql example: (required docker & docker compose)
make example-mysql

# to run the postgres example: (required docker & docker compose)
make example-postgres
```

## Built-in operators
Out of the box QueryFilter comes with a few operators built-in. However adding your
own custom operators is quite trivial. See examples in [operator.go](./operator.go).

The built-in operators are:
| `op` name       | SQL equivalent			   | Notes						   |
|-----------------|----------------------------|-------------------------------|
| `eq`            | `=`						   |							   |
| `gt`            | `>`						   |							   |
| `gte`           | `>=`					   |							   |
| `lt`            | `<`						   |							   |
| `lte`           | `<=`					   |							   |
| `in`            | `IN(?)`					   | Works on slices/arrays        |
| `not-in`        | `NOT IN(?)`                | works on slices/arrays        |
| `between`       | `BETWEEN ? AND ?`          | Works on slices/arrays of length 2|
| `is-null`       | `IS NULL` / `IS NOT NULL`  | Works on boolean types. Uses null/not null when passing true/false respectively|
| `not-null`      | `IS NOT NULL` / `IS NULL`  | Works on boolean types. Uses not null/null when passing true/false respectively|

## Other commands

```console
make test           # run the testsuite once
make watch          # run tests continually (watch mode; requires gotestsum)
make coverage       # generate test coverage report
```

## License

[MIT](./LICENSE)
