module github.com/tmw/queryfilter/examples/mysql

go 1.20

replace github.com/tmw/queryfilter => ../../

require (
	github.com/go-sql-driver/mysql v1.6.0
	github.com/jmoiron/sqlx v1.3.5
	github.com/tmw/queryfilter v0.0.0-00010101000000-000000000000
)

require github.com/mattn/go-sqlite3 v1.14.17 // indirect
