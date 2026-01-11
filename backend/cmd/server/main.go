package main

import (
	"log/slog"
	"os"

	"github.com/f1-rivals-cup/backend/internal/auth"
	"github.com/f1-rivals-cup/backend/internal/config"
	"github.com/f1-rivals-cup/backend/internal/database"
	"github.com/f1-rivals-cup/backend/internal/handler"
	custommiddleware "github.com/f1-rivals-cup/backend/internal/middleware"
	"github.com/f1-rivals-cup/backend/internal/repository"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Initialize database
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	permissionHistoryRepo := repository.NewPermissionHistoryRepository(db)
	leagueRepo := repository.NewLeagueRepository(db)
	participantRepo := repository.NewParticipantRepository(db)
	matchRepo := repository.NewMatchRepository(db)
	matchResultRepo := repository.NewMatchResultRepository(db)
	teamRepo := repository.NewTeamRepository(db)

	// Initialize JWT service
	jwtService := auth.NewJWTService(cfg.JWTSecret, cfg.JWTAccessExpiry, cfg.JWTRefreshExpiry)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler()
	authHandler := handler.NewAuthHandler(userRepo, jwtService)
	adminHandler := handler.NewAdminHandler(userRepo, permissionHistoryRepo)
	leagueHandler := handler.NewLeagueHandler(leagueRepo)
	participantHandler := handler.NewParticipantHandler(participantRepo, leagueRepo)
	matchHandler := handler.NewMatchHandler(matchRepo, leagueRepo)
	matchResultHandler := handler.NewMatchResultHandler(matchResultRepo, matchRepo, leagueRepo)
	teamHandler := handler.NewTeamHandler(teamRepo, leagueRepo)

	// Initialize Echo
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Health check endpoint
	e.GET("/health", healthHandler.Check)

	// API v1 routes
	v1 := e.Group("/api/v1")

	// Auth routes
	authGroup := v1.Group("/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/logout", authHandler.Logout)
	authGroup.POST("/refresh", authHandler.RefreshToken)
	authGroup.POST("/password-reset", authHandler.RequestPasswordReset)
	authGroup.POST("/password-reset/confirm", authHandler.ConfirmPasswordReset)

	// Admin routes (protected - require STAFF or ADMIN role)
	adminGroup := v1.Group("/admin")
	adminGroup.Use(custommiddleware.AuthMiddleware(jwtService))
	adminGroup.Use(custommiddleware.RequireRole(auth.RoleStaff, auth.RoleAdmin))

	// User management routes
	adminGroup.GET("/users", adminHandler.ListUsers, custommiddleware.RequirePermission(auth.PermUserView))
	adminGroup.GET("/users/:id", adminHandler.GetUser, custommiddleware.RequirePermission(auth.PermUserView))
	adminGroup.GET("/users/:id/history", adminHandler.GetUserPermissionHistory, custommiddleware.RequirePermission(auth.PermUserView))
	adminGroup.PUT("/users/:id/role", adminHandler.UpdateUserRole, custommiddleware.RequirePermission(auth.PermUserRoleChange))
	adminGroup.PUT("/users/:id/permissions", adminHandler.UpdateUserPermissions, custommiddleware.RequirePermission(auth.PermUserPermissionEdit))

	// Permission meta info
	adminGroup.GET("/permissions", adminHandler.GetPermissionsList, custommiddleware.RequirePermission(auth.PermUserView))
	adminGroup.GET("/permission-history", adminHandler.GetRecentPermissionHistory, custommiddleware.RequirePermission(auth.PermUserView))
	adminGroup.GET("/stats", adminHandler.GetUserStats, custommiddleware.RequirePermission(auth.PermUserView))

	// League routes (protected)
	adminGroup.POST("/leagues", leagueHandler.Create)
	adminGroup.GET("/leagues", leagueHandler.List)
	adminGroup.GET("/leagues/:id", leagueHandler.Get)
	adminGroup.PUT("/leagues/:id", leagueHandler.Update)
	adminGroup.DELETE("/leagues/:id", leagueHandler.Delete)

	// Admin participant routes
	adminGroup.GET("/leagues/:id/participants", participantHandler.ListByLeague)
	adminGroup.PUT("/participants/:id/status", participantHandler.UpdateStatus)
	adminGroup.PUT("/participants/:id/team", participantHandler.UpdateTeam)

	// Admin match routes
	adminGroup.POST("/leagues/:id/matches", matchHandler.Create)
	adminGroup.PUT("/matches/:id", matchHandler.Update)
	adminGroup.DELETE("/matches/:id", matchHandler.Delete)

	// Admin match result routes
	adminGroup.PUT("/matches/:id/results", matchResultHandler.BulkUpdate)
	adminGroup.DELETE("/matches/:id/results", matchResultHandler.Delete)

	// Admin team routes
	adminGroup.POST("/leagues/:id/teams", teamHandler.Create)
	adminGroup.PUT("/teams/:id", teamHandler.Update)
	adminGroup.DELETE("/teams/:id", teamHandler.Delete)

	// Public league routes
	leagueGroup := v1.Group("/leagues")
	leagueGroup.GET("", leagueHandler.List)
	leagueGroup.GET("/:id", leagueHandler.Get)
	leagueGroup.GET("/:id/matches", matchHandler.List)
	leagueGroup.GET("/:id/standings", matchResultHandler.Standings)
	leagueGroup.GET("/:id/teams", teamHandler.List)

	// Public match routes
	matchGroup := v1.Group("/matches")
	matchGroup.GET("/:id", matchHandler.Get)
	matchGroup.GET("/:id/results", matchResultHandler.List)

	// League participation routes (protected)
	leagueGroup.Use(custommiddleware.OptionalAuthMiddleware(jwtService))
	leagueGroup.GET("/:id/my-status", participantHandler.GetMyStatus)

	protectedLeagueGroup := v1.Group("/leagues")
	protectedLeagueGroup.Use(custommiddleware.AuthMiddleware(jwtService))
	protectedLeagueGroup.POST("/:id/join", participantHandler.Join)
	protectedLeagueGroup.DELETE("/:id/join", participantHandler.Cancel)

	// User profile routes (protected)
	meGroup := v1.Group("/me")
	meGroup.Use(custommiddleware.AuthMiddleware(jwtService))
	meGroup.GET("/participations", participantHandler.ListMyParticipations)

	// Start server
	slog.Info("Starting server", "port", cfg.ServerPort)
	if err := e.Start(":" + cfg.ServerPort); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
