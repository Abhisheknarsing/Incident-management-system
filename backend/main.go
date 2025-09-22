package main

import (
	"log"
	"net/http"
	"time"

	"incident-management-system/internal/database"
	"incident-management-system/internal/errors"
	"incident-management-system/internal/handlers"
	"incident-management-system/internal/logging"
	"incident-management-system/internal/monitoring"
	"incident-management-system/internal/services"
	"incident-management-system/internal/storage"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize logging
	logConfig := &logging.Config{
		Level:      logging.LevelInfo,
		Format:     "json",
		Output:     "stdout",
		AddSource:  true,
		TimeFormat: "2006-01-02T15:04:05.000Z",
	}

	if err := logging.InitGlobalLogger(logConfig); err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}

	logger := logging.GetGlobalLogger()
	logger.Info("Starting Incident Management System")

	// Initialize monitoring
	monitoring.InitMonitoring(logger)

	// Initialize memory monitoring
	memConfig := &monitoring.MemoryConfig{
		CollectionInterval: 30 * time.Second,
	}
	memMonitor := monitoring.NewMemoryMonitor(logger, memConfig)
	memMonitor.Start()
	defer memMonitor.Stop()

	// Initialize database
	dbConfig := &database.Config{
		DatabasePath: "incident_management.db",
	}
	db, err := database.NewDB(dbConfig)
	if err != nil {
		logger.Fatal("Failed to initialize database", err)
	}

	// Initialize database schema
	if err := db.InitializeDatabase(); err != nil {
		logger.Fatal("Failed to initialize database schema", err)
	}

	defer db.Close()

	// Initialize file storage
	fileStore := storage.NewFileStore("uploads")

	// Initialize services
	processingService := services.NewProcessingService(db.GetConnection(), fileStore)

	// Initialize handlers
	uploadHandler := handlers.NewUploadHandler(db.GetConnection(), fileStore, processingService)
	analyticsHandler := handlers.NewAnalyticsHandler(db.GetConnection())

	// Initialize Gin router with custom mode
	gin.SetMode(gin.ReleaseMode) // Disable Gin's default logging
	r := gin.New()

	// Add middleware
	r.Use(logging.RequestIDMiddleware())
	r.Use(logging.LoggingMiddleware(logger))
	r.Use(errors.RecoveryHandler())
	r.Use(errors.ErrorHandler())

	// Configure CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:5173"} // Vite dev server
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"}
	r.Use(cors.New(corsConfig))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		health := monitoring.GetHealthStatus()
		c.JSON(http.StatusOK, health)
	})

	// Monitoring endpoints
	r.GET("/metrics", func(c *gin.Context) {
		metrics, err := monitoring.ExportMetrics()
		if err != nil {
			errors.SendError(c, errors.InternalServer("Failed to export metrics"))
			return
		}
		c.Data(http.StatusOK, "application/json", metrics)
	})

	// Memory monitoring endpoints
	r.GET("/memory", func(c *gin.Context) {
		memUsage := memMonitor.GetMemoryUsage()
		c.JSON(http.StatusOK, memUsage)
	})

	r.POST("/memory/gc", func(c *gin.Context) {
		memMonitor.ForceGC()
		c.JSON(http.StatusOK, gin.H{"message": "Garbage collection forced"})
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

		// Analytics endpoints
		analytics := api.Group("/analytics")
		{
			// Timeline endpoints
			analytics.GET("/timeline/daily", analyticsHandler.GetDailyTimeline)
			analytics.GET("/timeline/weekly", analyticsHandler.GetWeeklyTimeline)
			analytics.GET("/timeline/overview", analyticsHandler.GetTimelineOverview)

			// Trend analysis endpoints
			analytics.GET("/trends", analyticsHandler.GetTrendAnalysis)

			// Metrics endpoints
			analytics.GET("/metrics/daily", analyticsHandler.GetTicketsPerDayMetrics)
			analytics.GET("/metrics/weekly", analyticsHandler.GetTicketsPerWeekMetrics)

			// Priority and Application Analysis endpoints
			analytics.GET("/priority", analyticsHandler.GetPriorityAnalysis)
			analytics.GET("/applications", analyticsHandler.GetApplicationAnalysis)
			analytics.GET("/resolution", analyticsHandler.GetResolutionAnalysis)
			analytics.GET("/performance", analyticsHandler.GetPerformanceMetrics)

			// Sentiment and Automation Analysis endpoints
			analytics.GET("/sentiment", analyticsHandler.GetSentimentAnalysis)
			analytics.GET("/automation", analyticsHandler.GetAutomationAnalysis)
			analytics.GET("/automation/reporting", analyticsHandler.GetITProcessAutomationReporting)
			analytics.GET("/summary", analyticsHandler.GetAnalyticsSummary)
		}
	}

	logger.Info("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		logger.Fatal("Failed to start server", err)
	}
}
