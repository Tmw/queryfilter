package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	qf "github.com/tmw/queryfilter"
)

type Filters struct {
	MinAge         *int       `filter:"age,op=gte"`
	MaxAge         *int       `filter:"age,op=lte"`
	FavouriteColor *string    `filter:"color"`
	Days           *[]string  `filter:"days,op=in"`
	After          *time.Time `filter:"due,op=gt"`
}

type Vehicle struct {
	Age   int
	Color string
	Days  string
	Due   time.Time
}

func setupDatabase() *sql.DB {
	db, err := sql.Open("sqlite3", "./local.db")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS vehicles (
		age INTEGER,
		color VARCHAR(255),
		days VARCHAR(255),
		due DATE
	);`)

	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`INSERT INTO vehicles (age, color, days, due) VALUES(
		23, "yellow", "monday", '2023-10-12'
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

	today := time.Now()

	f := Filters{
		MinAge:         &minAge,
		MaxAge:         &maxAge,
		FavouriteColor: &color,
		Days:           &days,
		After:          &today,
	}

	// note on configurability:
	// qf takes some global options. Eg:
	// qf.DefaultPlaceholderStrategy = qf.PlaceholderStrategyDollar
	// But there's the ability to override one of the options on a per case basis. eg:
	query, vars, err := qf.ToSql(f, qf.WithPlaceholderStrategy(qf.PlaceholderStrategyDollar))
	if err != nil {
		log.Fatal(err)
	}

	query = fmt.Sprintf("SELECT * FROM vehicles WHERE %s", query)
	fmt.Println("query:", query)
	fmt.Printf("vars: %v\n", vars)

	res := db.QueryRow(query, vars...)
	var vehicle Vehicle
	err = res.Scan(&vehicle.Age, &vehicle.Color, &vehicle.Days, &vehicle.Due)
	fmt.Println()

	if err == sql.ErrNoRows {
		fmt.Println("No records found")
		return
	}

	if err != nil {
		panic(err)
	}

	fmt.Printf("got result: %+v\n", vehicle)
}
