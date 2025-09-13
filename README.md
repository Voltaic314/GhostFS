# 👻 GhostFS - A File System Simulator

<div align="center">

![GhostFS Banner](assets/GhostFS%20Image%201.png)

**A powerful file system emulator for testing migration tools, file sync applications, and cloud storage integrations without the overhead of real file systems or expensive APIs.**

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![DuckDB](https://img.shields.io/badge/DuckDB-SQL-FFF000?style=flat&logo=duckdb)](https://duckdb.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Version](https://img.shields.io/badge/Version-0.1-green.svg)](https://github.com/Voltaic314/GhostFS/releases)

</div>

## 🎯 What is GhostFS?

GhostFS is a **SQL-backed file system emulator** that mimics a cloud storage API like Dropbox for example. Instead of dealing with real files and folders, GhostFS creates a virtual file system stored in a DuckDB database that you can traverse, query, and manipulate through a REST API.

Perfect for:
- **Testing file migration tools** (like [ByteWave](https://github.com/ByteWaveProject)) without moving real data
- **Simulating massive file systems** with millions of files and folders
- **Prototyping cloud storage integrations** with controllable environments
- **Load testing** file system operations at scale

## 🚀 Why GhostFS?

### The Problem
- Testing file migration tools requires **terabytes of real data**
- Cloud APIs have **rate limits** and **costs** during development
- Creating realistic folder structures manually is **time-consuming**
- Real file systems are **slow** for large-scale testing

### The Solution
- **Instant file system generation** with configurable depth and complexity
- **No storage overhead** - millions of "files" in a lightweight database
- **Full API control** - simulate network issues, auth failures, rate limits
- **Realistic testing** without the infrastructure costs

## ✨ Features

### Current (v0.1)
- 🗄️ **DuckDB Backend** - Fast, embedded SQL database
- 🌱 **Intelligent Seeding** - Generate realistic folder structures
- 🔄 **Multi-FS Mode** - Primary + secondary tables for migration testing
- 🎲 **Probabilistic Subsets** - Secondary tables with configurable `dst_prob`
- 📡 **REST API** - Standard HTTP endpoints for file operations
- 📊 **Batch Operations** - Create/delete multiple items at once
- 🎯 **Table Management** - List and manage multiple file systems

### Coming Soon (v0.2+)
- 🌐 **Network Simulation** - Configurable latency, jitter, timeouts
- 🔐 **Auth Simulation** - Token expiration, permission failures
- ⚡ **Rate Limiting** - Simulate API throttling
- 📈 **Metrics & Analytics** - Track usage patterns
- 🔧 **Plugin System** - Extend with custom behaviors

## 🏗️ Architecture & File System Modes

### Single-FS vs Multi-FS Mode

GhostFS operates in two distinct modes:

#### 🔵 **Single-FS Mode** (Default)
- Uses only the **primary table** (`nodes`)
- Perfect for basic file system testing
- All items exist in one unified file system

#### 🟡 **Multi-FS Mode** (Advanced)
- Uses **primary table + secondary tables**
- Simulates **source → destination** migration scenarios
- Secondary tables contain **probabilistic subsets** of the primary table
- Each item has a `dst_prob` chance of appearing in secondary tables

### How Secondary Tables Work

When generating a file system in Multi-FS mode:

1. **Primary Table** is populated with the complete file system
2. **Secondary Tables** are populated by iterating through primary items
3. Each item has a **probabilistic chance** (based on `dst_prob`) to be included
4. Results in **realistic migration scenarios** with missing files/folders

**Example with `dst_prob: 0.7`:**
```
Primary Table (Source):        Secondary Table (Destination):
├── folder1/                   ├── folder1/              ✅ (70% chance - included)
│   ├── file1.txt              │   ├── file1.txt         ✅ (70% chance - included)
│   ├── file2.txt              │   └── file3.txt         ✅ (70% chance - included)
│   └── file3.txt              └── folder3/              ✅ (70% chance - included)
├── folder2/                       └── file6.txt         ✅ (70% chance - included)
│   └── file4.txt              
├── folder3/                   ❌ folder2/ missing (30% chance - excluded)
│   ├── file5.txt              ❌ file2.txt missing (30% chance - excluded)
│   └── file6.txt              ❌ file4.txt missing (30% chance - excluded)
└── folder4/                   ❌ file5.txt missing (30% chance - excluded)
    └── file7.txt              ❌ folder4/ missing (30% chance - excluded)
```

### System Architecture

```
┌─────────────────┐    ┌──────────────┐    ┌─────────────────────┐
│   REST API      │────│   GhostFS    │────│      DuckDB         │
│  (Chi Router)   │    │   Server     │    │    ┌─────────────┐  │
└─────────────────┘    └──────────────┘    │    │ Primary     │  │
         │                       │         │    │ Table       │  │
         │              ┌────────▼────────┐ │    │ (nodes)     │  │
         │              │ Table Manager   │ │    └─────────────┘  │
         │              │ (Multi-table)   │ │    ┌─────────────┐  │
         │              └─────────────────┘ │    │ Secondary   │  │
         │                       │         │    │ Table 1     │  │
         │              ┌────────▼────────┐ │    │ (subset)    │  │
         │              │  Write Queue    │◄┤    └─────────────┘  │
         │              │  (Batching)     │ │    ┌─────────────┐  │
         │              └─────────────────┘ │    │ Secondary   │  │
         │                                 │    │ Table N     │  │
    ┌────▼─────┐                           │    │ (subset)    │  │
    │  Client  │                           │    └─────────────┘  │
    │   App    │                           └─────────────────────┘
    └──────────┘
```

## 🛠️ Installation & Setup

### Prerequisites
- Go 1.24.2 or higher
- Git

### Quick Start

```bash
# Clone the repository
git clone https://github.com/Voltaic314/GhostFS.git
cd GhostFS

# Install dependencies
go mod download

# Seed the database with sample data
go run main.go

# Start the API server
cd api
go run main.go server.go

# Server starts on http://localhost:8086 (configurable via config.json)
```

### Configuration

Create or modify `config.json`:

#### Single-FS Mode (Basic)
```json
{
  "database": {
    "path": "GhostFS.db",
    "tables": {
      "primary": {
        "table_name": "nodes",
        "min_child_folders": 2,
        "max_child_folders": 8,
        "min_child_files": 5,
        "max_child_files": 15,
        "min_depth": 3,
        "max_depth": 6
      }
    }
  },
  "network": {
    "address": "localhost",
    "port": 8086
  }
}
```

#### Multi-FS Mode (Migration Testing)
```json
{
  "database": {
    "path": "GhostFS.db",
    "tables": {
      "primary": {
        "table_name": "nodes_source",
        "min_child_folders": 3,
        "max_child_folders": 10,
        "min_child_files": 8,
        "max_child_files": 20,
        "min_depth": 4,
        "max_depth": 8
      },
      "secondary": {
        "destination_partial": {
          "table_name": "nodes_dest_partial",
          "dst_prob": 0.7
        },
        "destination_sparse": {
          "table_name": "nodes_dest_sparse", 
          "dst_prob": 0.3
        }
      }
    }
  },
  "network": {
    "address": "localhost",
    "port": 8086
  }
}
```

**Configuration Explained:**
- **`dst_prob: 0.7`** = 70% of items from primary will appear in this secondary table
- **`dst_prob: 0.3`** = 30% of items from primary will appear in this secondary table
- Multiple secondary tables simulate different migration scenarios

## 📚 API Reference

### Base URL: `http://localhost:8086`

### Tables Management

#### List All File Systems
```http
POST /tables/list
```

**Response:**
```json
{
  "success": true,
  "tables": [
    {
      "table_id": "uuid-here",
      "table_name": "nodes",
      "type": "primary"
    }
  ]
}
```

### File System Operations

#### List Items in Folder
```http
POST /items/list
Content-Type: application/json

{
  "table_id": "uuid-here",
  "folder_id": "root-folder-id"
}
```

#### Create Multiple Items
```http
POST /items/new
Content-Type: application/json

{
  "table_id": "uuid-here",
  "parent_id": "parent-folder-id",
  "items": [
    {"name": "New Folder", "type": "folder"},
    {"name": "document.txt", "type": "file", "size": 1024}
  ]
}
```

#### Delete Multiple Items
```http
POST /items/delete
Content-Type: application/json

{
  "table_id": "uuid-here",
  "item_ids": ["item-id-1", "item-id-2"]
}
```

#### Get Download URLs
```http
POST /items/download
Content-Type: application/json

{
  "table_id": "uuid-here",
  "file_ids": ["file-id-1", "file-id-2"]
}
```

## 🎮 Usage Examples

### Migration Testing Scenarios

#### Scenario 1: Incomplete Migration Detection
```bash
# 1. Generate source file system (primary table)
go run main.go

# 2. List source file system
curl -X POST http://localhost:8086/items/list \
  -d '{"table_id": "source-table-id", "folder_id": "root"}'

# 3. List destination file system (secondary table with dst_prob: 0.7)
curl -X POST http://localhost:8086/items/list \
  -d '{"table_id": "dest-table-id", "folder_id": "root"}'

# 4. Compare results - ~30% of files should be missing from destination
# Your migration tool should detect these missing files
```

#### Scenario 2: Incremental Sync Validation
```go
// Test your sync tool's ability to detect missing files
sourceItems := ghostfs.ListItems("source-table-id", "root")
destItems := ghostfs.ListItems("dest-partial-table-id", "root") 

// Your sync logic should identify missing items
missingItems := findMissingItems(sourceItems, destItems)
// With dst_prob: 0.7, expect ~30% missing items

// Run your incremental sync
syncTool.SyncMissing(missingItems)

// Validate sync completed successfully
```

### Testing a File Migration Tool

```go
// Connect to GhostFS
client := ghostfs.NewClient("http://localhost:8086")

// List available file systems
tables, _ := client.ListTables()
tableID := tables[0].TableID

// Get root folder contents
items, _ := client.ListItems(tableID, "root")

// Simulate migrating files
for _, item := range items {
    if item.Type == "file" {
        // Your migration logic here
        downloadURL, _ := client.GetDownloadURL(tableID, item.ID)
        // Process file...
    }
}
```

### Testing Rclone Integration

```bash
# Use GhostFS as a WebDAV endpoint (coming soon)
rclone sync ghostfs:/ local:backup/ --dry-run

# Or use the REST API directly
curl -X POST http://localhost:8086/items/list \
  -H "Content-Type: application/json" \
  -d '{"table_id": "your-table-id", "folder_id": "root"}'
```

## 🔧 Development

### Project Structure
```
GhostFS/
├── api/                    # REST API server
│   ├── routes/
│   │   ├── tables/        # Table management endpoints
│   │   └── items/         # File/folder CRUD endpoints
│   ├── main.go           # API server entry point
│   └── server.go         # Server configuration
├── db/                   # Database layer
│   ├── tables/          # Table management
│   └── write_queue.go   # Batched writes
├── seed/                # Database seeding
├── types/               # Shared types
├── config.json         # Configuration
└── main.go            # Seeder entry point
```

### Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 🎯 Use Cases

- **ByteWave** - File migration testing and validation
- **Cloud Storage SDKs** - Integration testing
- **Backup Tools** - Restore process validation
- **File Sync Apps** - Conflict resolution testing
- **Performance Testing** - Large-scale file operation benchmarks

## 🗺️ Roadmap

- [ ] **v0.2** - Network simulation (latency, failures)
- [ ] **v0.3** - Authentication simulation
- [ ] **v0.4** - Rate limiting and quotas
- [ ] **v0.5** - WebDAV/S3 protocol support
- [ ] **v1.0** - Plugin system and custom behaviors

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Support & Community

- 📖 **Documentation**: [Wiki](https://github.com/Voltaic314/GhostFS/wiki)
- 🐛 **Issues**: [GitHub Issues](https://github.com/Voltaic314/GhostFS/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/Voltaic314/GhostFS/discussions)

---

<div align="center">

**Built with ❤️ for the file migration and sync testing community**

[⭐ Star this repo](https://github.com/Voltaic314/GhostFS) if you find it useful!

</div>
