package main

import (
	"log"
	"os"
	"strings"

	"github.com/bmathus/pnregistry-webapi/api"
	"github.com/bmathus/pnregistry-webapi/internal/pn_registry"
	"github.com/gin-gonic/gin"
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
	// request routings
	pn_registry.AddRoutes(engine)
	engine.GET("/openapi", api.HandleOpenApi)
	engine.Run(":" + port)
}
