# Queryfilter

Construct parameterized SQL `where` clauses using Go structs.

## Installing

```golang
import (
    qf "github.com/tmw/queryfilter"
)
```

```console
go get github.com/tmw/queryfilter
```

## Usage

t.b.d

## Running examples

```console
# to run the sqlite example:
make example-sqlite

# to run the mysql example: (required docker & docker compose)
make example-mysql

# to run the postgres example: (required docker & docker compose)
make example-postgres
```

## Other commands

```console
make test           # run the testsuite once
make watch          # run tests continually (watch mode; requires gotestsum)
make coverage       # generate test coverage report
```

## License

[MIT](./LICENSE)
