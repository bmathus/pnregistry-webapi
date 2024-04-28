package main

import (
	"log"
	"os"
	"strings"

	"github.com/bmathus/pnregistry-webapi/api"
	"github.com/bmathus/pnregistry-webapi/internal/pn_registry"
	"github.com/gin-gonic/gin"

	"context"
	"time"

	"github.com/bmathus/pnregistry-webapi/internal/db_service"
	"github.com/gin-contrib/cors"
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
	// request routings
	pn_registry.AddRoutes(engine)
	engine.GET("/openapi", api.HandleOpenApi)
	engine.Run(":" + port)
}
