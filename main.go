package main 

import (
	"github.com/Voltaic314/GhostFS/api"
	"github.com/Voltaic314/GhostFS/db/seed"
)

func main() {
	seed.Seed()
	api.StartServer()
}