package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mindflow/agri-platform/internal/config"
	"github.com/mindflow/agri-platform/pkg/handlers"
	"github.com/mindflow/agri-platform/pkg/middleware"
	"github.com/mindflow/agri-platform/pkg/models"
	"github.com/mindflow/agri-platform/pkg/services"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := models.InitDB(cfg.DSN())
	if err != nil {
		log.Fatalf("init db: %v", err)
	}

	queueSvc := services.NewQueueService(100, 4)

	// Background context for goroutines
	bgCtx, bgCancel := context.WithCancel(context.Background())

	// Start background services
	taskSvc := services.NewTaskService(db)
	taskSvc.StartOverdueChecker(bgCtx)

	capacitySvc := services.NewCapacityService(db)
	capacitySvc.StartCapacityMonitor(bgCtx)

	router := setupRouter(db, cfg, queueSvc, taskSvc, capacitySvc)

	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("server starting on :%s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	bgCancel()
	queueSvc.Shutdown()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced shutdown: %v", err)
	}
	log.Println("server exited")
}

func setupRouter(db *gorm.DB, cfg *config.Config, queueSvc *services.QueueService, taskSvc *services.TaskService, capacitySvc *services.CapacityService) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Global audit middleware
	r.Use(middleware.AuditMiddleware(db))

	r.GET("/health", healthHandler(db))

	// Services
	authSvc := services.NewAuthService(db, cfg.JWTSecret)
	plotSvc := services.NewPlotService(db)
	deviceSvc := services.NewDeviceService(db)
	metricSvc := services.NewMetricService(db)
	monitorSvc := services.NewMonitorService(db, queueSvc)
	monDataSvc := services.NewMonitoringDataService(db, queueSvc)
	dashSvc := services.NewDashboardService(db)
	analysisSvc := services.NewAnalysisService(db)
	convSvc := services.NewConversationService(db)
	chatSvc := services.NewChatService(db)
	resultSvc := services.NewResultService(db)

	// Handlers
	authH := handlers.NewAuthHandler(authSvc)
	plotH := handlers.NewPlotHandler(plotSvc)
	deviceH := handlers.NewDeviceHandler(deviceSvc)
	metricH := handlers.NewMetricHandler(metricSvc)
	monitorH := handlers.NewMonitorHandler(monitorSvc, queueSvc)
	monDataH := handlers.NewMonitoringDataHandler(monDataSvc, queueSvc)
	dashH := handlers.NewDashboardHandler(dashSvc)
	analysisH := handlers.NewAnalysisHandler(analysisSvc)
	convH := handlers.NewConversationHandler(convSvc)
	taskH := handlers.NewTaskHandler(taskSvc)
	chatH := handlers.NewChatHandler(chatSvc)
	resultH := handlers.NewResultHandler(resultSvc)
	capacityH := handlers.NewCapacityHandler(capacitySvc)

	// Auth routes (public)
	v1 := r.Group("/v1")
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
		auth.GET("/me", middleware.AuthMiddleware(authSvc), authH.Me)
	}

	// Protected routes
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(authSvc))
	{
		// Plots
		plots := protected.Group("/plots")
		{
			plots.POST("", plotH.Create)
			plots.GET("", plotH.List)
			plots.GET("/:id", plotH.Get)
			plots.PUT("/:id", plotH.Update)
			plots.DELETE("/:id", plotH.Delete)
		}

		// Devices
		devices := protected.Group("/devices")
		{
			devices.POST("", deviceH.Create)
			devices.GET("", deviceH.List)
			devices.GET("/:id", deviceH.Get)
			devices.PUT("/:id", deviceH.Update)
			devices.DELETE("/:id", deviceH.Delete)
		}

		// Metrics
		metrics := protected.Group("/metrics")
		{
			metrics.POST("", metricH.Create)
			metrics.POST("/batch", metricH.BatchCreate)
			metrics.GET("", metricH.List)
			metrics.GET("/:id", metricH.Get)
			metrics.DELETE("/:id", metricH.Delete)
		}

		// Monitoring (device health & alerts)
		monitor := protected.Group("/monitor")
		{
			monitor.POST("/device", monitorH.CheckDevice)
			monitor.POST("/threshold", monitorH.ThresholdCheck)
			monitor.GET("/jobs/:id", monitorH.JobStatus)
			monitor.GET("/queue/status", monitorH.QueueStats)
			monitor.GET("/alerts", monitorH.ListAlerts)
			monitor.PATCH("/alerts/:id/resolve", monitorH.ResolveAlert)
		}

		// Monitoring Data (batch ingest, queries, aggregation, curves, trends, export)
		monitoring := protected.Group("/monitoring")
		{
			monitoring.POST("/ingest", monDataH.BatchIngest)
			monitoring.GET("/data", monDataH.List)
			monitoring.GET("/data/:id", monDataH.Get)
			monitoring.POST("/aggregate", monDataH.Aggregate)
			monitoring.POST("/curve", monDataH.RealtimeCurve)
			monitoring.POST("/trends", monDataH.Trends)
			monitoring.GET("/export/json", monDataH.ExportJSON)
			monitoring.GET("/export/csv", monDataH.ExportCSV)
			monitoring.GET("/jobs/:id", monDataH.JobStatus)
		}

		// Dashboards
		dashboards := protected.Group("/dashboards")
		{
			dashboards.POST("", dashH.Create)
			dashboards.GET("", dashH.List)
			dashboards.GET("/:id", dashH.Get)
			dashboards.PUT("/:id", dashH.Update)
			dashboards.DELETE("/:id", dashH.Delete)
		}

		// Analysis (trends, funnels, retention with drill-down)
		analysis := protected.Group("/analysis")
		{
			analysis.POST("/trends", analysisH.Trends)
			analysis.POST("/funnels", analysisH.Funnels)
			analysis.POST("/retention", analysisH.Retention)
		}

		// Orders & Conversations (Phase 7)
		orders := protected.Group("/orders")
		{
			orders.POST("", convH.CreateOrder)
			orders.GET("", convH.ListOrders)
			orders.GET("/:id", convH.GetOrder)
			orders.POST("/:id/messages", convH.PostMessage)
			orders.GET("/:id/messages", convH.ListMessages)
			orders.PATCH("/:id/messages/:msg_id/read", convH.MarkRead)
			orders.POST("/:id/transfer", convH.TransferTicket)
			orders.POST("/:id/templates/:template_id", convH.SendTemplate)
		}

		// Templates
		templates := protected.Group("/templates")
		{
			templates.POST("", convH.CreateTemplate)
			templates.GET("", convH.ListTemplates)
		}

		// Tasks (Phase 9)
		tasks := protected.Group("/tasks")
		{
			tasks.POST("", taskH.Create)
			tasks.POST("/generate", taskH.Generate)
			tasks.GET("", taskH.List)
			tasks.GET("/:id", taskH.Get)
			tasks.PUT("/:id", taskH.Update)
			tasks.DELETE("/:id", taskH.Delete)
			tasks.PATCH("/:id/submit", taskH.Submit)
			tasks.PATCH("/:id/review", taskH.Review)
			tasks.PATCH("/:id/complete", taskH.Complete)
		}

		// Chat (legacy simple messaging)
		chat := protected.Group("/chat")
		{
			chat.POST("", chatH.Send)
			chat.GET("", chatH.List)
			chat.PATCH("/:id/read", chatH.MarkRead)
		}

		// Results (Phase 8)
		results := protected.Group("/results")
		{
			results.POST("", resultH.Create)
			results.GET("", resultH.List)
			results.GET("/:id", resultH.Get)
			results.PUT("/:id", resultH.Update)
			results.DELETE("/:id", resultH.Delete)
			results.PATCH("/:id/transition", resultH.Transition)
			results.POST("/:id/notes", resultH.AppendNotes)
			results.POST("/:id/invalidate", resultH.Invalidate)
			results.POST("/field-rules", resultH.CreateFieldRule)
			results.GET("/field-rules", resultH.ListFieldRules)
		}

		// System (Phase 10)
		system := protected.Group("/system")
		{
			system.GET("/capacity", capacityH.CheckDisk)
			system.GET("/notifications", capacityH.ListNotifications)
		}
	}

	return r
}

func healthHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := models.Ping(db); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	}
}
