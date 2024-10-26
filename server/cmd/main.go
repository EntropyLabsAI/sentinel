package main

import (
	"log"
	"os"

	sentinel "github.com/entropylabsai/sentinel/server"
	"github.com/entropylabsai/sentinel/server/database"
)

func main() {
	db, err := database.NewPostgresqlStore(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	sentinel.InitAPI(db)
}
