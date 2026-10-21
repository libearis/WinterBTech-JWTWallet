package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"jwt-auth-wallet/internal/auth"
	"jwt-auth-wallet/internal/config"
	"jwt-auth-wallet/internal/database"
	"jwt-auth-wallet/internal/middleware"
	"jwt-auth-wallet/internal/user"
	"jwt-auth-wallet/internal/wallet"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := database.Connect(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to database and applied migrations successfully")

	jwtMgr := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTTTL, time.Now)

	authRepo := auth.NewPostgresRepository(db.Pool)
	authSvc := auth.NewService(authRepo, jwtMgr, cfg.BcryptCost, cfg.InitialBalance)
	authHandler := auth.NewHandler(authSvc)

	userSvc := user.NewService()
	userHandler := user.NewHandler(userSvc)

	walletRepo := wallet.NewPostgresRepository(db.Pool)
	walletSvc := wallet.NewService(walletRepo, time.Now)
	walletHandler := wallet.NewHandler(walletSvc)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(middleware.CORS(cfg.CORSAllowedOrigins))

	// Health check handlers
	healthHandler := func(c *gin.Context) {
		pingCtx, pingCancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer pingCancel()

		if err := db.Pool.Ping(pingCtx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "message": "database unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}

	router.GET("/healthz", healthHandler)

	api := router.Group("/api")
	{
		api.GET("/healthz", healthHandler)

		// Auth public endpoints
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/refresh", middleware.Auth(jwtMgr), authHandler.Refresh)
		}

		// Protected endpoints
		protected := api.Group("/")
		protected.Use(middleware.Auth(jwtMgr))
		{
			protected.GET("/me", userHandler.Me)

			walletGroup := protected.Group("/wallet")
			{
				walletGroup.GET("", walletHandler.GetBalance)
				walletGroup.GET("/transactions", walletHandler.GetTransactions)
				walletGroup.POST("/transfers", walletHandler.Transfer)
			}
		}
	}

	addr := fmt.Sprintf(":%s", cfg.AppPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting API server on %s (env: %s)", addr, cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server ListenAndServe error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down API server gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
