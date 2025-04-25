package database

import (
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB

func Setup(dsn string) {
	var err error
	DB, err = sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalln(err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatalln(err)
	}

	schema, err := os.ReadFile("sql/init.sql")
	if err != nil {
		log.Fatalln(err)
	}

	DB.MustExec(string(schema))

}
