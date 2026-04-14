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

	router := setupRouter(db, cfg)

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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced shutdown: %v", err)
	}
	log.Println("server exited")
}

func setupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/health", healthHandler(db))

	// Services
	authSvc := services.NewAuthService(db, cfg.JWTSecret)
	plotSvc := services.NewPlotService(db)
	deviceSvc := services.NewDeviceService(db)
	metricSvc := services.NewMetricService(db)

	// Handlers
	authH := handlers.NewAuthHandler(authSvc)
	plotH := handlers.NewPlotHandler(plotSvc)
	deviceH := handlers.NewDeviceHandler(deviceSvc)
	metricH := handlers.NewMetricHandler(metricSvc)

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
