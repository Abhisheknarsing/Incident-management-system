package main

import (
	"log"
	"net/http"

	"incident-management-system/internal/database"
	"incident-management-system/internal/handlers"
	"incident-management-system/internal/services"
	"incident-management-system/internal/storage"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	dbConfig := &database.Config{
		DatabasePath: "incident_management.db",
	}
	db, err := database.NewDB(dbConfig)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize file storage
	fileStore := storage.NewFileStore("uploads")

	// Initialize services
	processingService := services.NewProcessingService(db.GetConnection(), fileStore)

	// Initialize handlers
	uploadHandler := handlers.NewUploadHandler(db.GetConnection(), fileStore, processingService)

	// Initialize Gin router
	r := gin.Default()

	// Configure CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:5173"} // Vite dev server
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(corsConfig))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"message": "Incident Management System API",
		})
	})

	// API routes
	api := r.Group("/api")
	{
		// Upload endpoints
		api.POST("/uploads", uploadHandler.UploadFile)
		api.GET("/uploads", uploadHandler.GetUploads)
		api.GET("/uploads/:id", uploadHandler.GetUpload)
		api.POST("/uploads/:id/process", uploadHandler.ProcessUpload)
		api.GET("/uploads/:id/status", uploadHandler.GetProcessingStatus)
	}

	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}