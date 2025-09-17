package main

import (
	"github.com/Voltaic314/GhostFS/code/api"
	// "github.com/Voltaic314/GhostFS/code/db/seed"
)

func main() {
	cfgPath := "config.json"
	// seed.InitDB(cfgPath)
	api.StartServer(cfgPath)
}
