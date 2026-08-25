package service

import (
	"context"
	"testing"

	productv1 "storemesh-product-service/gen/storemesh/product/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCatalogCRUD(t *testing.T) {
	catalog := NewCatalog()
	created, err := catalog.CreateProduct(context.Background(), &productv1.CreateProductRequest{Product: &productv1.Product{Sku: "SKU-1", Name: "Coffee", Currency: "USD", PriceMinor: 1299}})
	if err != nil {
		t.Fatalf("create product: %v", err)
	}
	if created.GetProduct().GetId() == "" {
		t.Fatal("expected generated product ID")
	}
	got, err := catalog.GetProduct(context.Background(), &productv1.GetProductRequest{Id: created.GetProduct().GetId()})
	if err != nil || got.GetProduct().GetSku() != "SKU-1" {
		t.Fatalf("get product: got=%v err=%v", got, err)
	}
	_, err = catalog.ArchiveProduct(context.Background(), &productv1.ArchiveProductRequest{Id: created.GetProduct().GetId()})
	if err != nil {
		t.Fatalf("archive product: %v", err)
	}
}

func TestCatalogRejectsInvalidProduct(t *testing.T) {
	_, err := NewCatalog().CreateProduct(context.Background(), &productv1.CreateProductRequest{Product: &productv1.Product{Name: "Missing SKU"}})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected invalid argument, got %v", err)
	}
}
