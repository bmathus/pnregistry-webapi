package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bmathus/pnregistry-webapi/api"
	"github.com/bmathus/pnregistry-webapi/internal/db_service"
	"github.com/bmathus/pnregistry-webapi/internal/pn_registry"
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
	dbService := db_service.NewMongoService[pn_registry.Record](db_service.MongoServiceConfig{})
	defer dbService.Disconnect(context.Background())
	engine.Use(func(ctx *gin.Context) {
		ctx.Set("db_service", dbService)
		ctx.Next()
	})

	// register custom validators for patientId,fullname,employer and reason fields
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("only-digits-max-length-10", pn_registry.PatientIDValidator)
		v.RegisterValidation("max-length-50", pn_registry.MaxLengthValidator)
		v.RegisterValidation("not-valid-reason-value", pn_registry.ReasonValidator)
	}

	// request routings
	pn_registry.AddRoutes(engine)
	engine.GET("/openapi", api.HandleOpenApi)
	engine.Run(":" + port)
}
