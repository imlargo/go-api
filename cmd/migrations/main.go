package main

import (
	"log"

	"github.com/imlargo/go-api/internal/config"
	postgres "github.com/imlargo/go-api/internal/database"
)

func main() {
	cfg := config.LoadConfig()

	db, err := postgres.NewPostgres(cfg.Database.URL)
	if err != nil {
		log.Fatal("Could not initialize database: ", err)
	}

	err = postgres.Migrate(db)
	if err != nil {
		log.Fatal("Could not run migrations: ", err)
		return
	}
}
