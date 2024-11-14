package handlers

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/bmathus/pnregistry-webapi/internal/db_service"
	"github.com/bmathus/pnregistry-webapi/internal/pn_registry"
	"github.com/bmathus/pnregistry-webapi/internal/pn_registry/models"
	"github.com/bmathus/pnregistry-webapi/internal/pn_registry/utils"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var testDbService db_service.DbService[models.Record]

func TestMain(m *testing.M) {

	// Initialize dbService for tests
	testDbService = db_service.NewMongoService[models.Record](db_service.MongoServiceConfig{
		ServerHost: "localhost",
		ServerPort: 27018,
		UserName:   "root",
		Password:   "neUhaDnes",
		DbName:     "pn-registry-test",
		Collection: "record",
	})

	if testDbService == nil {
		log.Fatal("testDbService failed to initialize in TestMain")
	}

	// // Clean up any records before running tests
	if err := testDbService.ClearCollection(context.Background()); err != nil {
		log.Fatalf("Failed to clear test database: %v", err)
	}

	code := m.Run()

	// Disconnect after tests finish
	if err := testDbService.ClearCollection(context.Background()); err != nil {
		log.Fatalf("Failed to clear test database: %v", err)
	}
	if err := testDbService.Disconnect(context.Background()); err != nil {
		log.Fatalf("Failed to disconnect from test database: %v", err)
	}

	os.Exit(code)
}

func SetupTestEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.Use(func(ctx *gin.Context) {
		ctx.Set("db_service", testDbService)
		ctx.Next()
	})

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("only-digits-max-length-10", utils.PatientIDValidator)
		v.RegisterValidation("max-length-50", utils.MaxLengthValidator)
		v.RegisterValidation("not-valid-reason-value", utils.ReasonValidator)
	}

	pn_registry.AddRoutes(engine)
	return engine
}
