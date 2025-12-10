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

// withCORS replica la configuración usada en products-api para habilitar requests desde el frontend local.
func withCORS(next http.Handler) http.Handler {
	allowedOrigins := map[string]struct{}{
		"http://localhost:5173": {},
		"http://127.0.0.1:5173": {},
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if _, ok := allowedOrigins[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Vary", "Origin")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	cfg := config.Load()
	log.Printf("users-api config: port=%s db_host=%s db_port=%s db_name=%s jwt_expiration=%dmin", cfg.ServerPort, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.JWTExpirationMinutes)

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
	mux.Handle("/users/", authMiddleware(http.HandlerFunc(userHandler.GetUserByID)))

	addr := ":" + cfg.ServerPort
	log.Printf("users-api listening on %s", addr)

	handler := middleware.RequestLogger(withCORS(mux))

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
