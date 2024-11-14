package utils

import (
	"fmt"

	"github.com/bmathus/pnregistry-webapi/internal/db_service"
	"github.com/bmathus/pnregistry-webapi/internal/pn_registry/models"
	"github.com/gin-gonic/gin"
)

func GetDatabaseService(ctx *gin.Context) (db_service.DbService[models.Record], string, error) {
	value, exists := ctx.Get("db_service")
	if !exists {
		return nil, "db not found", fmt.Errorf("db not found")
	}

	db, ok := value.(db_service.DbService[models.Record])
	if !ok {
		return nil, "db_service context is not of type db_service.DbService", fmt.Errorf("cannot cast db_service context to db_service.DbService")
	}

	return db, "", nil
}
