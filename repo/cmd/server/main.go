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

	// Deterministic SQL migration bootstrap: applies any SQL file under
	// cfg.MigrationsDir that has not yet been recorded in
	// schema_migrations. This guarantees partitioning/archive schema is in
	// place on clean environments without a manual step.
	if err := models.RunSQLMigrations(db, cfg.MigrationsDir); err != nil {
		log.Fatalf("run sql migrations: %v", err)
	}

	queueSvc := services.NewQueueService(100, 4)

	// Background context for goroutines
	bgCtx, bgCancel := context.WithCancel(context.Background())

	// Start background services
	taskSvc := services.NewTaskService(db)
	taskSvc.StartOverdueChecker(bgCtx)

	capacitySvc := services.NewCapacityService(db)
	capacitySvc.StartCapacityMonitor(bgCtx)

	retentionSvc := services.NewRetentionService(db)
	retentionSvc.StartRetentionWorker(bgCtx)

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
	authSvc := services.NewAuthService(db, cfg.JWTSecret, cfg.EncryptionKey)
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
	indicatorSvc := services.NewIndicatorService(db)

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
	indicatorH := handlers.NewIndicatorHandler(indicatorSvc)

	// Role shortcuts
	adminOnly := middleware.RoleGuard("admin")
	adminOrResearcher := middleware.RoleGuard("admin", "researcher")
	adminResearcherReviewer := middleware.RoleGuard("admin", "researcher", "reviewer")
	adminResearcherReviewerCS := middleware.RoleGuard("admin", "researcher", "reviewer", "customer_service")

	// Auth routes (public)
	v1 := r.Group("/v1")
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
		auth.GET("/me", middleware.AuthMiddleware(authSvc), authH.Me)
	}

	// Protected routes (require authentication)
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(authSvc))
	{
		// Plots — admin/researcher can create/update/delete; all authenticated can read
		plots := protected.Group("/plots")
		{
			plots.POST("", adminOrResearcher, plotH.Create)
			plots.GET("", plotH.List)
			plots.GET("/:id", plotH.Get)
			plots.PUT("/:id", adminOrResearcher, plotH.Update)
			plots.DELETE("/:id", adminOrResearcher, plotH.Delete)
		}

		// Devices — admin/researcher can manage; all authenticated can read
		devices := protected.Group("/devices")
		{
			devices.POST("", adminOrResearcher, deviceH.Create)
			devices.GET("", deviceH.List)
			devices.GET("/:id", deviceH.Get)
			devices.PUT("/:id", adminOrResearcher, deviceH.Update)
			devices.DELETE("/:id", adminOrResearcher, deviceH.Delete)
		}

		// Metrics — admin/researcher can manage; all authenticated can read
		metrics := protected.Group("/metrics")
		{
			metrics.POST("", adminOrResearcher, metricH.Create)
			metrics.POST("/batch", adminOrResearcher, metricH.BatchCreate)
			metrics.GET("", metricH.List)
			metrics.GET("/:id", metricH.Get)
			metrics.DELETE("/:id", adminOrResearcher, metricH.Delete)
		}

		// Monitoring (device health & alerts) — admin/researcher can manage
		monitor := protected.Group("/monitor")
		{
			monitor.POST("/device", adminOrResearcher, monitorH.CheckDevice)
			monitor.POST("/threshold", adminOrResearcher, monitorH.ThresholdCheck)
			monitor.GET("/jobs/:id", monitorH.JobStatus)
			monitor.GET("/queue/status", monitorH.QueueStats)
			monitor.GET("/alerts", monitorH.ListAlerts)
			monitor.PATCH("/alerts/:id/resolve", adminOrResearcher, monitorH.ResolveAlert)
		}

		// Monitoring Data — admin/researcher can ingest; all authenticated can query
		monitoring := protected.Group("/monitoring")
		{
			monitoring.POST("/ingest", adminOrResearcher, monDataH.BatchIngest)
			monitoring.GET("/data", monDataH.List)
			monitoring.GET("/data/:id", monDataH.Get)
			monitoring.POST("/aggregate", monDataH.Aggregate)
			monitoring.POST("/curve", monDataH.RealtimeCurve)
			monitoring.POST("/trends", monDataH.Trends)
			monitoring.GET("/export/json", monDataH.ExportJSON)
			monitoring.GET("/export/csv", monDataH.ExportCSV)
			monitoring.GET("/jobs/:id", monDataH.JobStatus)
		}

		// Dashboards — all authenticated users (user-scoped at service level)
		dashboards := protected.Group("/dashboards")
		{
			dashboards.POST("", dashH.Create)
			dashboards.GET("", dashH.List)
			dashboards.GET("/:id", dashH.Get)
			dashboards.PUT("/:id", dashH.Update)
			dashboards.DELETE("/:id", dashH.Delete)
		}

		// Analysis — admin/researcher/reviewer can run analyses
		analysis := protected.Group("/analysis")
		analysis.Use(adminResearcherReviewer)
		{
			analysis.POST("/trends", analysisH.Trends)
			analysis.POST("/funnels", analysisH.Funnels)
			analysis.POST("/retention", analysisH.Retention)
		}

		// Indicators (version management) — admin/researcher can manage; all authenticated can read
		indicators := protected.Group("/indicators")
		{
			indicators.POST("", adminOrResearcher, indicatorH.Create)
			indicators.GET("", indicatorH.List)
			indicators.GET("/:id", indicatorH.Get)
			indicators.PUT("/:id", adminOrResearcher, indicatorH.Update)
			indicators.DELETE("/:id", adminOnly, indicatorH.Delete)
			indicators.GET("/:id/versions", indicatorH.ListVersions)
			indicators.GET("/:id/versions/:version", indicatorH.GetVersion)
		}

		// Orders & Conversations — admin/researcher/reviewer/customer_service
		orders := protected.Group("/orders")
		orders.Use(adminResearcherReviewerCS)
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

		// Templates — admin/customer_service can manage
		templates := protected.Group("/templates")
		{
			templates.POST("", middleware.RoleGuard("admin", "customer_service"), convH.CreateTemplate)
			templates.GET("", convH.ListTemplates)
		}

		// Tasks — admin/researcher can create/update/delete; reviewer can review/complete
		tasks := protected.Group("/tasks")
		{
			tasks.POST("", adminOrResearcher, taskH.Create)
			tasks.POST("/generate", adminOrResearcher, taskH.Generate)
			tasks.GET("", taskH.List)
			tasks.GET("/:id", taskH.Get)
			tasks.PUT("/:id", adminOrResearcher, taskH.Update)
			tasks.DELETE("/:id", adminOrResearcher, taskH.Delete)
			tasks.PATCH("/:id/submit", adminOrResearcher, taskH.Submit)
			tasks.PATCH("/:id/review", adminResearcherReviewer, taskH.Review)
			tasks.PATCH("/:id/complete", adminResearcherReviewer, taskH.Complete)
		}

		// Chat (legacy simple messaging) — all authenticated
		chat := protected.Group("/chat")
		{
			chat.POST("", chatH.Send)
			chat.GET("", chatH.List)
			chat.PATCH("/:id/read", chatH.MarkRead)
		}

		// Results — admin/researcher can create/update/delete; admin/researcher/reviewer for transitions
		results := protected.Group("/results")
		{
			results.POST("", adminOrResearcher, resultH.Create)
			results.GET("", resultH.List)
			results.GET("/:id", resultH.Get)
			results.PUT("/:id", adminOrResearcher, resultH.Update)
			results.DELETE("/:id", adminOrResearcher, resultH.Delete)
			results.PATCH("/:id/transition", adminResearcherReviewer, resultH.Transition)
			results.POST("/:id/notes", adminResearcherReviewer, resultH.AppendNotes)
			results.POST("/:id/invalidate", adminOnly, resultH.Invalidate)
			results.POST("/field-rules", adminOnly, resultH.CreateFieldRule)
			results.GET("/field-rules", resultH.ListFieldRules)
		}

		// System — admin only
		system := protected.Group("/system")
		system.Use(adminOnly)
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
