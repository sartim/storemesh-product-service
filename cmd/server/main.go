package main

import (
	"context"
	"log"
	"net"
	"os"

	"storemesh-product-service/internal/auth"
	productv1 "storemesh-product-service/gen/storemesh/product/v1"
	"storemesh-product-service/internal/repository"
	"storemesh-product-service/internal/service"
	"google.golang.org/grpc"
)

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}
	serverOptions := []grpc.ServerOption{}
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		issuer, audience := os.Getenv("JWT_ISSUER"), os.Getenv("JWT_AUDIENCE")
		if issuer == "" { issuer = "storemesh-product-service" }
		if audience == "" { audience = "storemesh-platform" }
		serverOptions = append(serverOptions, grpc.UnaryInterceptor(auth.UnaryInterceptor(secret, issuer, audience)))
	}
	server := grpc.NewServer(serverOptions...)
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		store, err := repository.OpenProductStore(context.Background(), databaseURL)
		if err != nil { log.Fatal(err) }
		defer store.Close()
		productv1.RegisterProductCatalogServiceServer(server, service.NewPersistentCatalog(store))
	} else {
		productv1.RegisterProductCatalogServiceServer(server, service.NewCatalog())
	}
	log.Println("product service listening on :50051")
	if err := server.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
