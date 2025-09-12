package tables

// PrimaryTableConfig represents configuration for the primary table
type PrimaryTableConfig struct {
	TableName       string `json:"table_name"`
	Seed            int64  `json:"seed,omitempty"`
	MinChildFolders int    `json:"min_child_folders,omitempty"`
	MaxChildFolders int    `json:"max_child_folders,omitempty"`
	MinChildFiles   int    `json:"min_child_files,omitempty"`
	MaxChildFiles   int    `json:"max_child_files,omitempty"`
	MinDepth        int    `json:"min_depth,omitempty"`
	MaxDepth        int    `json:"max_depth,omitempty"`
}

// SecondaryTableConfig represents configuration for a secondary table
type SecondaryTableConfig struct {
	TableName string  `json:"table_name"`
	DstProb   float64 `json:"dst_prob"` // Probability of placing node in this table (0.0-1.0)
}

// TestConfig represents the configuration for test harness
type TestConfig struct {
	Database struct {
		Path   string `json:"path"`
		Tables struct {
			Primary   PrimaryTableConfig              `json:"primary"`
			Secondary map[string]SecondaryTableConfig `json:"secondary"` // map of table ID to config
		} `json:"tables"`
	} `json:"database"`
	Network struct {
		Address string `json:"address"`
		Port    int    `json:"port"`
	} `json:"network"`
}
