package main

import (
	"log"
	"os"

	asteroid "github.com/asteroidai/asteroid/server"
	database "github.com/asteroidai/asteroid/server/db"
)

func main() {
	db, err := database.NewPostgresqlStore(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	asteroid.InitAPI(db)
}
