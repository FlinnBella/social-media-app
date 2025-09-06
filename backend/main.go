package main

import (
	"log"
	"social-media-ai-video/config"
	"social-media-ai-video/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// requireAPIKey returns a middleware that validates the X-API-Key or Bearer token
func requireAPIKey(cfg *config.APIConfig) gin.HandlerFunc {
	return func(c *gin.Context) {

		if cfg.Environment == "development" {
			c.Next()
			return
		} else {
			if cfg.APIKey == "" {
				c.AbortWithStatusJSON(401, gin.H{"status": "error", "error": "API key not configured"})
				return
			}
			key := c.GetHeader("X-API-Key")
			if key == "" {
				// Fallback to Authorization: Bearer
				auth := c.GetHeader("Authorization")
				const prefix = "Bearer "
				if len(auth) > len(prefix) && auth[:len(prefix)] == prefix {
					key = auth[len(prefix):]
				}
			}
			if key != cfg.APIKey {
				c.AbortWithStatusJSON(401, gin.H{"status": "error", "error": "invalid API key"})
				return
			}
			c.Next()
		}
	}
}

func main() {
	// Load configuration
	cfg := config.LoadAPIConfig()

	// Initialize Gin router
	r := gin.Default()

	// Configure CORS for localhost development
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-Key"},
		AllowCredentials: true,
	}))

	// Initialize handlers
	videoHandler := handlers.NewVideoHandler(cfg)

	// API routes
	api := r.Group("/api")
	{
		api.POST("/generate-video-pexels", videoHandler.GenerateVideoPexels)
		api.POST("/generate-video-reels", videoHandler.GenerateVideoReels)
		api.POST("/generate-video-pro-reels", videoHandler.GenerateProReels)
		//add requireAPIKey middleware later
		//api.GET("/composition", videoHandler.GetComposition)
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
