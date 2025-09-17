package main

import (
	"fmt"
	"log"

	"github.com/Voltaic314/GhostFS/code/sdk"
)

func main() {
	// Option 1: Simple initialization - auto-discovers GhostFS.db and config.json
	// Fails if database doesn't exist (safe default)
	client, err := sdk.NewGhostFSClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Option 2: Auto-discovery with database generation
	// Uncomment the lines below to enable auto-generation of database
	// client, err := sdk.NewGhostFSClient(true)
	// if err != nil {
	//     log.Fatal(err)
	// }
	// defer client.Close()

	// List all available tables
	tables, err := client.ListTables()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Available tables: %d\n", len(tables))
	for _, table := range tables {
		fmt.Printf("  - %s (%s): %s\n", table.TableID, table.Type, table.TableName)
	}

	// Get the root node for the primary table
	root, err := client.GetRoot("primary")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Root node: %s (%s) at %s\n", root.Name, root.Type, root.Path)

	// List items in the root folder
	items, err := client.ListItems("primary", root.ID, false)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d items in root folder:\n", len(items))
	for _, item := range items {
		fmt.Printf("  - %s (%s)\n", item.Name, item.Type)
	}

	// Show cache statistics
	stats := client.GetCacheStats()
	fmt.Printf("Cache stats: %+v\n", stats)
}
