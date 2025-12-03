package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
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

	productService := services.NewProductService(productRepo, publisher)
	productHandler := handlers.NewProductHandler(productService)
	purchaseService := services.NewPurchaseService(productRepo, purchaseRepo, publisher)
	purchaseHandler := handlers.NewPurchaseHandler(purchaseService)
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
	mux.Handle("/compras", handlers.MethodHandler{
		Post: authMiddleware(http.HandlerFunc(purchaseHandler.CreatePurchase)),
	})
	mux.Handle("/compras/mias", handlers.MethodHandler{
		Get: authMiddleware(http.HandlerFunc(purchaseHandler.ListMyPurchases)),
	})

	addr := ":" + cfg.ServerPort
	log.Printf("products-api listening on %s", addr)
	if err := http.ListenAndServe(addr, middleware.RequestLogger(mux)); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
