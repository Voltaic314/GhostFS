# GhostFS SDK

A simple, easy-to-use SDK for ByteWave and other applications to interact with GhostFS.

## Quick Start

```go
import "github.com/Voltaic314/GhostFS/code/sdk"

// Initialize with config file
client, err := sdk.NewGhostFSClient("config.json")
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// List items in a folder
items, err := client.ListItems("primary", "folder-id", false)
if err != nil {
    log.Fatal(err)
}

for _, item := range items {
    fmt.Printf("Found: %s (%s)\n", item.Name, item.Type)
}
```

## Initialization

### With Config File (Recommended)
```go
client, err := sdk.NewGhostFSClient("config.json")
```
- Uses the specified config file
- Database path is determined from config (or current directory if not specified)
- If database doesn't exist and `generate_if_not_exists` is true, creates a new one
- Perfect for ByteWave integration

### Auto-Discovery with Default Config
```go
client, err := sdk.NewGhostFSClient("")
```
- Looks for `config.json` in current directory
- Database path is determined from config (or current directory if not specified)
- If database doesn't exist and `generate_if_not_exists` is true, creates a new one

### Custom Database Path
```go
client, err := sdk.NewGhostFSClientWithDB("/path/to/GhostFS.db")
```
- Use a specific database file
- Still auto-discovers `config.json`
- Useful for testing with different databases

## API Methods

### ListItems
```go
items, err := client.ListItems(tableID, folderID, foldersOnly)
```
- `tableID`: Table to query (e.g., "primary", "secondary_0")
- `folderID`: ID of the folder to list
- `foldersOnly`: If true, only return folders; if false, return files and folders

### GetRoot
```go
root, err := client.GetRoot(tableID)
```
- Gets the root node for a specific table

### ListTables
```go
tables, err := client.ListTables()
```
- Lists all available tables with their IDs and types

### Cache Management
```go
// Get cache statistics
stats := client.GetCacheStats()

// Clear the in-memory cache
client.ClearCache()
```

## Performance

- **Sub-millisecond performance** for ByteWave stress testing
- **In-memory caching** of seeds and existence maps
- **Deterministic generation** - same results every time
- **No HTTP overhead** when used directly

## Error Handling

All methods return Go errors that should be checked:

```go
items, err := client.ListItems("primary", "folder-id", false)
if err != nil {
    // Handle error (invalid table ID, folder not found, etc.)
    log.Printf("Error: %v", err)
    return
}
```

## Example

See `example_usage.go` for a complete working example.

## Configuration

The SDK uses a simplified config file format. Create a `config.json` file:

```json
{
  "database": {
    "path": "GhostFS.db",
    "generate_if_not_exists": true,
    "tables": {
      "primary": {
        "table_name": "src_nodes",
        "seed": 12345,
        "min_child_folders": 2,
        "max_child_folders": 5,
        "min_child_files": 3,
        "max_child_files": 8,
        "min_depth": 1,
        "max_depth": 10
      },
      "secondary": {
        "dst_0": {
          "table_name": "dst_nodes",
          "dst_prob": 0.7
        }
      }
    }
  }
}
```

### Config Fields

- `database.path`: Path to database file (optional, defaults to current directory)
- `database.generate_if_not_exists`: Whether to create database if it doesn't exist
- `database.tables.primary`: Primary table configuration (required)
- `database.tables.secondary`: Secondary table configurations (optional)

## Requirements

- Configuration file (`config.json`)
- Database file (created automatically if `generate_if_not_exists` is true)
