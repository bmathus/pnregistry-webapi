package main

import (
	"log"
	"os"
	"strings"

	"github.com/bmathus/pnregistry-webapi/api"
	"github.com/bmathus/pnregistry-webapi/internal/pn_registry"
	"github.com/gin-gonic/gin"

	"context"
	"regexp"
	"time"

	"github.com/bmathus/pnregistry-webapi/internal/db_service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Custom validator for PatientId field
// Custom validator for PatientId field
func patientIDValidator(fl validator.FieldLevel) bool {
	patientID := fl.Field().String()
	matched, _ := regexp.MatchString(`^\d{1,10}$`, patientID)
	return matched
}

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

	// Register custom validator for patientId field
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("only-digits-max-length-10", patientIDValidator)
	}

	// request routings
	pn_registry.AddRoutes(engine)
	engine.GET("/openapi", api.HandleOpenApi)
	engine.Run(":" + port)
}
