package tables

import (
	"net/http"

	coreTables "github.com/Voltaic314/GhostFS/code/core/tables"
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

	// Call core logic
	coreResp, err := coreTables.ListTables(database)
	if err != nil {
		api.InternalError(w, err.Error())
		return
	}

	// Convert core response to API response
	responseData := ListTablesResponseData{Tables: coreResp.Tables}
	api.Success(w, responseData)
}
