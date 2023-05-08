package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	qf "github.com/tmw/queryfilter"
)

type Filters struct {
	MinAge         *int      `filter:"age,op=gte"`
	MaxAge         *int      `filter:"age,op=lte"`
	FavouriteColor *string   `filter:"color"`
	Days           *[]string `filter:"days,op=in"`
}

func setupDatabase() *sql.DB {
	db, err := sql.Open("sqlite3", "./local.db")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS vehicles (
		age INTEGER,
		color VARCHAR(255),
		days VARCHAR(255)
	);`)

	if err != nil {
		panic(err)
	}

	return db
}

func main() {
	db := setupDatabase()
	defer db.Close()

	minAge, maxAge, color := 18, 35, "yellow"
	days := []string{
		"monday", "wednesday",
	}

	f := Filters{
		MinAge:         &minAge,
		MaxAge:         &maxAge,
		FavouriteColor: &color,
		Days:           &days,
	}

	// note on configurability:
	// qf takes some global options. Eg:
	// qf.DefaultPlaceholderStrategy = qf.PlaceholderStrategyDollar
	// But there's the ability to override one of the options on a per case basis. eg:
	query, vars, err := qf.ToSql(f, qf.WithPlaceholderOffset(12), qf.WithPlaceholderStrategy(qf.PlaceholderStrategyDollar))
	if err != nil {
		log.Fatal(err)
	}

	query = fmt.Sprintf("SELECT * FROM vehicles WHERE %s", query)
	fmt.Println("query:", query)
	fmt.Printf("vars: %v\n", vars)

	res, err := db.Exec(query, vars...)
	if err != nil {
		panic(err)
	}

	fmt.Println(res)
}
