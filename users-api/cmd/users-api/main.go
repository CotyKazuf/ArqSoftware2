package main

import (
	"log"
	"net/http"

	"users-api/internal/config"
	"users-api/internal/database"
	"users-api/internal/handlers"
	"users-api/internal/middleware"
	"users-api/internal/repositories"
	"users-api/internal/services"
)

func main() {
	cfg := config.Load()

	db, err := database.Init(cfg)
	if err != nil {
		log.Fatalf("database init: %v", err)
	}

	userRepo := repositories.NewGormUserRepository(db)
	userService := services.NewUserService(userRepo, cfg.JWTSecret, cfg.JWTExpirationMinutes)

	if err := userService.EnsureAdminUser("Admin", cfg.AdminEmail, cfg.AdminDefaultPassword); err != nil {
		log.Printf("admin bootstrap failed: %v", err)
	}

	userHandler := handlers.NewUserHandler(userService)
	authMiddleware := middleware.AuthMiddleware(cfg.JWTSecret)

	mux := http.NewServeMux()
	mux.HandleFunc("/users/register", userHandler.Register)
	mux.HandleFunc("/users/login", userHandler.Login)
	mux.Handle("/users/me", authMiddleware(http.HandlerFunc(userHandler.Me)))

	addr := ":" + cfg.ServerPort
	log.Printf("users-api listening on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
