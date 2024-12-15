package main

import (
	"log"
	"os"

	asteroid "github.com/asteroidai/asteroid/server"
	database "github.com/asteroidai/asteroid/server/db"
)

func main() {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		log.Fatalf("DATABASE_URL is not set")
	}

	db, err := database.NewPostgresqlStore(url)
	if err != nil {
		log.Fatalf("Failed to connect to the database with URL %s: %v", url, err)
	}
	defer db.Close()

	asteroid.InitAPI(db)
}
