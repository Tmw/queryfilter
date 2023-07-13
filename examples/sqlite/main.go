package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	qf "github.com/tmw/queryfilter"
)

var (
	DatabaseName = "local.db"
	Schema       = `
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			status TEXT NOT NULL,
			story_points INTEGER NOT NULL,
			due_date DATETIME NOT NULL
		);
	`

	Seed = `
		INSERT OR IGNORE INTO tasks (id, title, status, story_points, due_date) VALUES
			(1, 'Code Review Workflow Management', 'done', 5, '2023-01-20'),
			(2, 'Agile Project Planning Tools', 'doing', 8, '2023-05-05'),
			(3, 'Continuous Integration Automation Framework', 'todo', 3, '2023-07-31'),
			(4, 'Version Control System Integration', 'done', 2, '2023-02-14'),
			(5, 'Test Case Management System', 'doing', 1, '2023-03-23');
	`
)

type Filters struct {
	Status    *[]string  `filter:"status,op=in"`
	MinPoints *int       `filter:"story_points,op=gte"`
	MaxPoints *int       `filter:"story_points,op=lte"`
	DueBefore *time.Time `filter:"due_date,op=gt"`
}

type Task struct {
	ID          int64
	Title       string
	Status      string
	StoryPoints int       `db:"story_points"`
	DueDate     time.Time `db:"due_date"`
}

func setupDatabase() *sqlx.DB {
	db := sqlx.MustConnect("sqlite3", DatabaseName)
	db.MustExec(Schema)
	db.MustExec(Seed)
	return db
}

func main() {
	db := setupDatabase()
	defer db.Close()

	statuses, minPoints, maxPoints := []string{"todo", "doing"}, 2, 5
	filter := Filters{
		Status:    &statuses,
		MinPoints: &minPoints,
		MaxPoints: &maxPoints,
	}

	query, vars, err := qf.ToSQL(filter)
	if err != nil {
		log.Fatal(err)
	}

	query = fmt.Sprintf("SELECT * FROM tasks WHERE %s LIMIT 1", query)
	fmt.Printf("query: %s\n", query)
	fmt.Printf("vars: %v\n", vars)

	var task Task
	if err := db.Get(&task, query, vars...); err != nil {
		log.Fatal(err)
	}

	if err == sql.ErrNoRows {
		fmt.Println("no records found")
		return
	}

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("result: %+v\n", task)
}
