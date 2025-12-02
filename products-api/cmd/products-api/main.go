package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"products-api/internal/config"
	"products-api/internal/database"
	"products-api/internal/handlers"
	"products-api/internal/middleware"
	"products-api/internal/rabbitmq"
	"products-api/internal/repositories"
	"products-api/internal/services"
)

func main() {
	cfg := config.Load()

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

	publisher, err := rabbitmq.NewPublisher(cfg)
	if err != nil {
		log.Fatalf("rabbitmq init: %v", err)
	}
	defer publisher.Close()

	productService := services.NewProductService(productRepo, publisher)
	productHandler := handlers.NewProductHandler(productService)
	authMiddleware := middleware.AuthMiddleware(cfg.JWTSecret)

	mux := http.NewServeMux()
	mux.Handle("/products", handlers.MethodHandler{
		Get:  http.HandlerFunc(productHandler.ListProducts),
		Post: authMiddleware(middleware.RequireAdmin(http.HandlerFunc(productHandler.CreateProduct))),
	})
	mux.Handle("/products/", handlers.MethodHandler{
		Get:    http.HandlerFunc(productHandler.GetProduct),
		Put:    authMiddleware(middleware.RequireAdmin(http.HandlerFunc(productHandler.UpdateProduct))),
		Delete: authMiddleware(middleware.RequireAdmin(http.HandlerFunc(productHandler.DeleteProduct))),
	})

	addr := ":" + cfg.ServerPort
	log.Printf("products-api listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
