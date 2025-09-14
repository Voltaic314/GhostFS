package tables

import (
	"net/http"

	"github.com/Voltaic314/GhostFS/code/db"
	"github.com/Voltaic314/GhostFS/code/db/tables"
	"github.com/Voltaic314/GhostFS/code/types/api"
	dbTypes "github.com/Voltaic314/GhostFS/code/types/db"
)

// Response struct for this endpoint
type ListTablesResponseData struct {
	Tables []dbTypes.TableInfo `json:"tables"`
}

// HandleListTables handles requests to list all node tables
func HandleListTables(w http.ResponseWriter, r *http.Request, serverInterface interface{}) {
	// Cast to the actual server type
	server := serverInterface.(interface {
		GetTableManager() *tables.TableManager
		GetDB() *db.DB
	})

	database := server.GetDB()

	// Get all table mappings with types from the database
	tableMappingsWithTypes, err := tables.GetAllTableMappingsWithTypes(database)
	if err != nil {
		api.InternalError(w, "Failed to retrieve table mappings from database")
		return
	}

	var tableList []dbTypes.TableInfo
	for tableID, info := range tableMappingsWithTypes {
		tableList = append(tableList, dbTypes.TableInfo{
			TableID:   tableID,
			TableName: info["table_name"],
			Type:      info["type"],
		})
	}

	// Return successful response
	responseData := ListTablesResponseData{Tables: tableList}
	api.Success(w, responseData)
}
