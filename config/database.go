package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() {
	var err error

	dsn := fmt.Sprintf(
		"host=db  user=%s password=%s dbname=%s port=5432 sslmode=disable", 
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
}