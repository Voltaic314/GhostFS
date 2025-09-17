package main

import (
	"fmt"
	"log"

	"github.com/Voltaic314/GhostFS/code/sdk"
)

func main() {
	// Option 1: Simple initialization with config file
	// Uses sdk_config.json in examples directory
	client, err := sdk.NewGhostFSClient("examples/sdk_config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Option 2: Auto-discovery with default config
	// Uncomment the lines below to use default config discovery
	// client, err := sdk.NewGhostFSClient("")
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

	primary_table_ID := tables[0].TableID

	// Get the root node for the primary table
	root, err := client.GetRoot(primary_table_ID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Root node:\n")
	fmt.Printf("  ID:        %s\n", root.ID)
	fmt.Printf("  Name:      %s\n", root.Name)
	fmt.Printf("  Type:      %s\n", root.Type)
	fmt.Printf("  Path:      %s\n", root.Path)
	fmt.Printf("  ParentID:  %s\n", root.ParentID)
	fmt.Printf("  Size:      %d\n", root.Size)
	fmt.Printf("  Level:     %d\n", root.Level)
	fmt.Printf("  Checked:   %v\n", root.Checked)
	fmt.Printf("  CreatedAt: %v\n", root.CreatedAt)
	fmt.Printf("  UpdatedAt: %v\n", root.UpdatedAt)

	// List items in the root folder
	items, err := client.ListItems(primary_table_ID, root.ID, false)
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
