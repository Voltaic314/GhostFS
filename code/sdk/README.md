# GhostFS SDK

A simple, easy-to-use SDK for ByteWave and other applications to interact with GhostFS.

## Quick Start

```go
import "github.com/Voltaic314/GhostFS/code/sdk"

// Auto-discover database and config
client, err := sdk.NewGhostFSClient()
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

### Auto-Discovery (Recommended)
```go
client, err := sdk.NewGhostFSClient()
```
- Automatically finds `GhostFS.db` in the project root (2 levels up from `code/sdk/`)
- Automatically finds `config.json` in the project root (2 levels up from `code/sdk/`)
- Fails if database doesn't exist (safe default)
- Perfect for ByteWave integration
- **Note**: Files are found relative to the package location, not your current working directory

### Auto-Discovery with Database Generation
```go
client, err := sdk.NewGhostFSClient(true)
```
- Automatically finds `GhostFS.db` in the project root (2 levels up from `code/sdk/`)
- **If no database exists, creates a new one with root folders**
- Automatically finds `config.json` in the project root (2 levels up from `code/sdk/`)
- ⚠️ **Use with caution** - this will create a new database if none exists

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

## Requirements

- GhostFS database file (`GhostFS.db`)
- Configuration file (`config.json`)
- Both files should be in the project root directory (2 levels up from `code/sdk/`)
- The SDK will automatically find these files relative to the package location, not your current working directory
