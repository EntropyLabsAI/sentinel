package main

import (
	sentinel "github.com/entropylabsai/sentinel/server"
	"github.com/entropylabsai/sentinel/server/memorystore"
)

func main() {
	sentinel.InitAPI(memorystore.New())
}
