package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"time"

	"products-api/internal/clients"
	"products-api/internal/config"
	"products-api/internal/database"
	"products-api/internal/handlers"
	"products-api/internal/middleware"
	"products-api/internal/rabbitmq"
	"products-api/internal/repositories"
	"products-api/internal/services"
)

// withCORS agrega los encabezados necesarios para permitir requests del frontend local.
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
	mongoHost := cfg.MongoURI
	if parsed, err := url.Parse(cfg.MongoURI); err == nil {
		mongoHost = parsed.Host
	}
	rabbitHost := cfg.RabbitMQURL
	if parsed, err := url.Parse(cfg.RabbitMQURL); err == nil {
		rabbitHost = parsed.Host
	}
	log.Printf("products-api config: port=%s mongo_host=%s mongo_db=%s rabbit_host=%s rabbit_exchange=%s", cfg.ServerPort, mongoHost, cfg.MongoDB, rabbitHost, cfg.RabbitMQExchange)

	mongoClient, mongoDB, err := database.InitMongo(cfg)
	if err != nil {
		log.Fatalf("mongo init: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Printf("mongo disconnect: %v", err)
		}
	}()

	collection := mongoDB.Collection("products")
	productRepo := repositories.NewMongoProductRepository(collection)
	purchaseCollection := mongoDB.Collection("purchases")
	purchaseRepo := repositories.NewMongoPurchaseRepository(purchaseCollection)

	publisher, err := rabbitmq.NewPublisher(cfg)
	if err != nil {
		log.Fatalf("rabbitmq init: %v", err)
	}
	defer publisher.Close()

	usersClient := clients.NewUsersClient(cfg.UsersAPIBaseURL)
	productService := services.NewProductService(productRepo, publisher, usersClient)
	productHandler := handlers.NewProductHandler(productService)
	purchaseService := services.NewPurchaseService(productRepo, purchaseRepo, publisher)
	purchaseHandler := handlers.NewPurchaseHandler(purchaseService)
	authMiddleware := middleware.AuthMiddleware(cfg.JWTSecret)

	mux := http.NewServeMux()
	mux.Handle("/products", handlers.MethodHandler{
		Get:  http.HandlerFunc(productHandler.ListProducts),
		Post: authMiddleware(http.HandlerFunc(productHandler.CreateProduct)),
	})
	mux.Handle("/products/", handlers.MethodHandler{
		Get:    http.HandlerFunc(productHandler.GetProduct),
		Put:    authMiddleware(http.HandlerFunc(productHandler.UpdateProduct)),
		Delete: authMiddleware(http.HandlerFunc(productHandler.DeleteProduct)),
	})
	mux.Handle("/compras", handlers.MethodHandler{
		Post: authMiddleware(http.HandlerFunc(purchaseHandler.CreatePurchase)),
	})
	mux.Handle("/compras/mias", handlers.MethodHandler{
		Get: authMiddleware(http.HandlerFunc(purchaseHandler.ListMyPurchases)),
	})

	addr := ":" + cfg.ServerPort
	log.Printf("products-api listening on %s", addr)

	// encadenamos: mux -> CORS -> logger
	handler := middleware.RequestLogger(withCORS(mux))

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
