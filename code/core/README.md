# GhostFS Core Package

This package contains the core business logic for GhostFS, extracted from the HTTP API handlers to make it importable by ByteWave and other applications.

## Architecture

The core package mirrors the API structure:

```
code/core/
├── items/
│   ├── list.go          # ListItems function
│   └── get_root.go      # GetRoot function
└── tables/
    └── list.go          # ListTables function
```

## Usage

### For ByteWave (Direct Import)

```go
import (
    "github.com/Voltaic314/GhostFS/code/core/items"
    "github.com/Voltaic314/GhostFS/code/core/tables"
)

// List items in a folder
req := items.ListItemsRequest{
    TableID:     "primary",
    FolderID:    "folder-id",
    FoldersOnly: false,
}

resp, err := items.ListItems(tableManager, database, generator, req)
if err != nil {
    log.Fatal(err)
}

for _, item := range resp.Items {
    fmt.Printf("Found item: %s (%s)\n", item.Name, item.Type)
}
```

### For HTTP API (Wrapper)

The HTTP API handlers now simply call these core functions and handle JSON serialization/deserialization.

## Benefits

1. **Performance**: ByteWave can import core directly for sub-millisecond performance
2. **Consistency**: Same logic used everywhere - no duplication
3. **Testability**: Core logic can be tested without HTTP overhead
4. **Flexibility**: Easy to add new interfaces (gRPC, WebSocket, etc.)

## Functions

### items.ListItems
Lists all items (files and folders) in a folder using deterministic generation.

### items.GetRoot
Gets the root node for a table.

### tables.ListTables
Lists all available node tables.

Each function takes the necessary dependencies (tableManager, database, generator) and returns structured responses with proper error handling.
