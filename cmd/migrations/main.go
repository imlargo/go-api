package main

import (
	"log"

	"github.com/nicolailuther/butter/internal/config"
	"github.com/nicolailuther/butter/internal/database"
)

func main() {
	cfg := config.LoadConfig()

	db, err := database.NewPostgres(cfg.Database.URL)
	if err != nil {
		log.Fatal("Error: Could not initialize database: ", err)
	}

	err = database.Migrate(db)
	if err != nil {
		log.Fatal("Could not run migrations: ", err)
		return
	}
}
