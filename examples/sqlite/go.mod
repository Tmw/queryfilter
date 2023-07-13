module github.com/tmw/queryfilter/examples/sqlite

go 1.20

replace github.com/tmw/queryfilter => ../../

require (
	github.com/jmoiron/sqlx v1.3.5
	github.com/mattn/go-sqlite3 v1.14.6
	github.com/tmw/queryfilter v0.0.0-00010101000000-000000000000
)
