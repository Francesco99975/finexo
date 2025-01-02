package database

import (
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var db *sqlx.DB

func Setup(dsn string) {
	var err error
	db, err = sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalln(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalln(err)
	}

	schema, err := os.ReadFile("sql/init.sql")
	if err != nil {
		log.Fatalln(err)
	}

	db.MustExec(string(schema))
}
