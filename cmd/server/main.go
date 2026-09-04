package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	productv1 "github.com/sartim/storemesh-product-service/gen/storemesh/product/v1"
	"github.com/sartim/storemesh-product-service/internal/auth"
	"github.com/sartim/storemesh-product-service/internal/observability"
	"github.com/sartim/storemesh-product-service/internal/repository"
	"github.com/sartim/storemesh-product-service/internal/service"
	"google.golang.org/grpc"
)

func main() {
	grpcAddr := env("GRPC_ADDR", ":50051")
	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal(err)
	}
	serverOptions := []grpc.ServerOption{}
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		issuer, audience := os.Getenv("JWT_ISSUER"), os.Getenv("JWT_AUDIENCE")
		if issuer == "" {
			issuer = "storemesh-product-service"
		}
		if audience == "" {
			audience = "storemesh-platform"
		}
		oidc, err := auth.NewOIDCValidator(os.Getenv("KEYCLOAK_ISSUER"), os.Getenv("KEYCLOAK_AUDIENCE"))
		if err != nil {
			log.Fatalf("configure Keycloak OIDC: %v", err)
		}
		serverOptions = append(serverOptions, grpc.UnaryInterceptor(auth.UnaryInterceptor(secret, issuer, audience, oidc)))
	}
	server := grpc.NewServer(serverOptions...)
	go serveMetrics(env("METRICS_ADDR", ":8080"))
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		store, err := repository.OpenProductStore(context.Background(), databaseURL)
		if err != nil {
			log.Fatal(err)
		}
		defer store.Close()
		productv1.RegisterProductCatalogServiceServer(server, service.NewPersistentCatalog(store))
	} else {
		productv1.RegisterProductCatalogServiceServer(server, service.NewCatalog())
	}
	log.Println("product service listening on " + grpcAddr)
	if err := server.Serve(listener); err != nil {
		log.Fatal(err)
	}
}

func serveMetrics(addr string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", observability.Handler())
	server := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Println("product metrics listening on " + addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("metrics server: %v", err)
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
