package main

import (
	"log"
	"net"

	productv1 "storemesh-product-service/gen/storemesh/product/v1"
	"storemesh-product-service/internal/service"
	"google.golang.org/grpc"
)

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}
	server := grpc.NewServer()
	productv1.RegisterProductCatalogServiceServer(server, service.NewCatalog())
	log.Println("product service listening on :50051")
	if err := server.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
