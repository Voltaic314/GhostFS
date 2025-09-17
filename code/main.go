package main

import (
	"log"
	"fmt"
	// "github.com/Voltaic314/GhostFS/code/api"
	// "github.com/Voltaic314/GhostFS/code/db/seed"
	"github.com/Voltaic314/GhostFS/code/sdk"
)

func main() {
	cfgPath := "config.json"
	// seed.InitDB(cfgPath)
	// api.StartServer(cfgPath)
	// Initialize with config file
	client, err := sdk.NewGhostFSClient(cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	tables, err := client.ListTables()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Available tables:", tables)

	table_ID := tables[0].TableID
	root, err := client.GetRoot(table_ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Root node:", root)

	items, err := client.ListItems(table_ID, root.ID, false)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Found", len(items), "items in root folder")
}
