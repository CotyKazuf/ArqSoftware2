package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"search-api/internal/cache"
	"search-api/internal/config"
	"search-api/internal/handlers"
	"search-api/internal/middleware"
	"search-api/internal/rabbitmq"
	"search-api/internal/services"
	"search-api/internal/solr"
)

func main() {
	cfg := config.Load()

	cacheTTL := time.Duration(cfg.CacheTTLSeconds) * time.Second
	memoryCache := cache.NewCCacheLayer(cfg.CacheMaxEntries)
	distributedCache := cache.NewMemcachedLayer(cfg.MemcachedAddr)
	layeredCache := cache.NewLayeredCache(memoryCache, distributedCache, cacheTTL)

	solrClient := solr.NewClient(cfg.SolrURL, cfg.SolrCore)
	searchService := services.NewSearchService(solrClient, layeredCache, cacheTTL)
	eventProcessor := services.NewEventProcessor(searchService)

	consumer, err := rabbitmq.NewConsumer(rabbitmq.ConsumerConfig{
		URL:      cfg.RabbitURL,
		Exchange: cfg.RabbitExchange,
		Queue:    cfg.RabbitQueue,
	}, eventProcessor)
	if err != nil {
		log.Fatalf("rabbitmq init: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		if err := consumer.Start(ctx); err != nil && err != context.Canceled {
			log.Printf("rabbit consumer stopped: %v", err)
		}
	}()
	defer consumer.Close()

	searchHandler := handlers.NewSearchHandler(searchService)
	authMiddleware := middleware.AuthMiddleware(cfg.JWTSecret)

	mux := http.NewServeMux()
	mux.Handle("/search/products", handlers.MethodHandler{Get: http.HandlerFunc(searchHandler.SearchProducts)})
	mux.Handle("/search/cache/flush", authMiddleware(middleware.RequireAdmin(http.HandlerFunc(searchHandler.FlushCache))))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: mux,
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutdown signal received")
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	log.Printf("search-api listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
