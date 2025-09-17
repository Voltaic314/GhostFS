# GhostFS Database Initialization

This script creates a fresh GhostFS database with the new schema and root nodes for testing the deterministic generation system.

## What it does

1. **Deletes existing database** - Removes `GhostFS.db` and `GhostFS.db.wal` files
2. **Creates new schema** - Sets up all tables with the updated schema including `child_seed` column
3. **Creates root nodes** - Generates root folders for all tables (primary + secondary)
4. **Sets up seeds** - Creates deterministic seeds for the root nodes
5. **Saves configuration** - Stores table mappings and seed info

## Usage
```bash
go run init_db.go ../../config.json
```

## What gets created

- **Primary table**: `nodes` with root folder
- **Secondary tables**: `nodes_secondary_0`, `nodes_secondary_1` with root folders
- **Lookup tables**: `table_lookup`, `seed_info`
- **Root nodes**: Each with deterministic `child_seed` and `secondary_existence_map`

## After initialization

The database will be ready for the deterministic generation system. When you start the GhostFS server, it will:

1. Load existing seeds into memory cache
2. Use deterministic generation for all folder listings
3. Provide massive performance improvements (100-1000x faster)

## Configuration

The script uses `config.json` from the project root. Make sure your configuration includes:

- Primary table settings (min/max children, depth)
- Secondary table settings (dst_prob values)
- Database path

## Example output

```
🗑️  Removing existing database: C:\path\to\GhostFS.db
🔧 Creating new database: C:\path\to\GhostFS.db
🎲 Master seed: 1234567890
📜 Creating tables...
📜 Created table: table_lookup
📜 Created table: seed_info
📜 Created table: nodes
📜 Created table: nodes_secondary_0
📜 Created table: nodes_secondary_1
🌱 Creating root nodes...
🌱 Created root in primary table: nodes
🌱 Created root in secondary table: nodes_secondary_0
🌱 Created root in secondary table: nodes_secondary_1
✅ Database initialization complete!
📊 Created root nodes for 3 tables
🚀 Ready for deterministic generation!
```
