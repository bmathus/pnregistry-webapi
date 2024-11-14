package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bmathus/pnregistry-webapi/internal/db_service"
	"github.com/bmathus/pnregistry-webapi/internal/pn_registry"
	"github.com/bmathus/pnregistry-webapi/internal/pn_registry/models"
	"github.com/bmathus/pnregistry-webapi/internal/pn_registry/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func main() {
	log.Printf("Server started")
	port := os.Getenv("PN_REGISTRY_API_PORT")
	if port == "" {
		port = "8080"
	}
	environment := os.Getenv("PN_REGISTRY_API_ENVIRONMENT")
	if !strings.EqualFold(environment, "production") { // case insensitive comparison
		gin.SetMode(gin.DebugMode)
	}
	engine := gin.New()
	engine.Use(gin.Recovery())

	// setup cors middleware
	corsMiddleware := cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "PUT", "POST", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{""},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	})
	engine.Use(corsMiddleware)

	// setup context update middleware
	dbService := db_service.NewMongoService[models.Record](db_service.MongoServiceConfig{})
	fmt.Println(dbService)
	defer dbService.Disconnect(context.Background())
	engine.Use(func(ctx *gin.Context) {
		if dbService == nil {
			log.Fatal("TestDbService is nil in SetupTestEngine")
		}
		ctx.Set("db_service", dbService)
		ctx.Next()
	})

	// register custom validators for patientId,fullname,employer and reason fields
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("only-digits-max-length-10", utils.PatientIDValidator)
		v.RegisterValidation("max-length-50", utils.MaxLengthValidator)
		v.RegisterValidation("not-valid-reason-value", utils.ReasonValidator)
	}

	// request routings
	pn_registry.AddRoutes(engine)
	engine.Run(":" + port)
}
