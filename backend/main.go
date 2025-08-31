package main

import (
	"log"
	"social-media-ai-video/config"
	"social-media-ai-video/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadAPIConfig()

	// Initialize Gin router
	r := gin.Default()

	// Configure CORS for localhost development
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	// Initialize handlers
	videoHandler := handlers.NewVideoHandler(cfg)

	// API routes
	api := r.Group("/api")
	{
		api.POST("/generate-video", videoHandler.GenerateVideo)
		api.GET("/composition", videoHandler.GetComposition)
	}

	// Serve static files from ./tmp at /static
	r.Static("/static", "./tmp")

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}