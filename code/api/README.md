# GhostFS API

A mock filesystem HTTP API for testing FS API operations. 
Applications include migration apps, backup & sync apps, and 
file system apps that need a library which can mock & mimmick real
filesystem APIs.

## Features

- **Batched List Operations**: Efficiently handles multiple concurrent folder listing requests
- **Path-Based Joins**: Uses database path joins for optimal performance
- **Mock File Downloads**: Generates random binary data for file downloads
- **Health Monitoring**: Built-in health check endpoint
- **Graceful Shutdown**: Proper cleanup on server termination

## API Endpoints

### Health Check
- `GET /health` - Returns server health status

### Filesystem Operations
- `GET /list/{path}` - List contents of a folder
- `GET /is-directory/{path}` - Check if path is a directory
- `POST /create-folder` - Create a new folder (not implemented)
- `POST /create-file` - Create a new file (not implemented)
- `GET /file/{fileID}/{filename}` - Get file download URL
- `GET /download/{fileID}/{filename}` - Download file content

## Usage

```bash
# Start the server
go run . -config config.json
```

## Configuration

The server reads configuration from the same JSON file used by the seeder. The new configuration supports flexible table management:

### Table Configuration

- **Primary Table**: Contains the main generation configuration and table name
- **Secondary Tables**: Optional additional tables with their own `dst_prob` values
- **Single Table Mode**: Set `secondary` to an empty object `{}` or omit it entirely
- **Multi Table Mode**: Define secondary tables with integer keys and table configurations

### Configuration Fields

**Primary Table:**
- `table_name`: Name of the table in the database
- `seed`: Random seed for generation (0 = use current time)
- `min_child_folders`/`max_child_folders`: Range for folder generation
- `min_child_files`/`max_child_files`: Range for file generation  
- `min_depth`/`max_depth`: Range for tree depth

**Secondary Tables:**
- `table_name`: Name of the table in the database
- `dst_prob`: Probability (0.0-1.0) of placing nodes in this table

```json
{
  "database": {
    "path": "GhostFS.db",
    "tables": {
      "primary": {
        "table_name": "nodes",
        "seed": 0,
        "min_child_folders": 1,
        "max_child_folders": 5,
        "min_child_files": 0,
        "max_child_files": 10,
        "min_depth": 3,
        "max_depth": 6
      },
      "secondary": {
        "0": {
          "table_name": "nodes_secondary_0",
          "dst_prob": 0.7
        },
        "1": {
          "table_name": "nodes_secondary_1", 
          "dst_prob": 0.5
        }
      }
    }
  },
  "network": {
    "address": "127.0.0.1",
    "port": 8086
  }
}
```

## Batching Strategy

The server uses a sophisticated batching system for list operations:

1. **Request Collection**: Multiple workers can request folder contents simultaneously
2. **Background Batching**: Requests are collected every 5ms or when batch size (10) is reached
3. **Single Query**: One database query with `UNION ALL` to get children from both tables
4. **Response Distribution**: Results are grouped by parent path and sent to respective workers

This approach provides:
- **High Concurrency**: Handles multiple workers efficiently
- **Database Efficiency**: Reduces query load with batching
- **Low Latency**: Fast response times for individual requests
