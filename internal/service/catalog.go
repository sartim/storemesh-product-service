package service

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/google/uuid"
	productv1 "github.com/sartim/storemesh-product-service/gen/storemesh/product/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var errInvalidProduct = errors.New("product must include sku, name, currency, and a non-negative price")

type Catalog struct {
	productv1.UnimplementedProductCatalogServiceServer
	mu       sync.RWMutex
	products map[string]*productv1.Product
}

func NewCatalog() *Catalog {
	return &Catalog{products: make(map[string]*productv1.Product)}
}

func (c *Catalog) CreateProduct(_ context.Context, req *productv1.CreateProductRequest) (*productv1.CreateProductResponse, error) {
	if req == nil || req.GetProduct() == nil {
		return nil, status.Error(codes.InvalidArgument, errInvalidProduct.Error())
	}
	product := clone(req.GetProduct())
	if err := validate(product); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, existing := range c.products {
		if existing.GetSku() == product.GetSku() {
			return nil, status.Errorf(codes.AlreadyExists, "product SKU %q already exists", product.GetSku())
		}
	}
	product.Id = uuid.NewString()
	product.Status = productv1.ProductStatus_PRODUCT_STATUS_ACTIVE
	product.CreatedAt = timestamppb.Now()
	product.UpdatedAt = product.CreatedAt
	c.products[product.Id] = product
	return &productv1.CreateProductResponse{Product: clone(product)}, nil
}

func (c *Catalog) GetProduct(_ context.Context, req *productv1.GetProductRequest) (*productv1.GetProductResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	product, ok := c.products[req.GetId()]
	if !ok {
		return nil, status.Error(codes.NotFound, "product not found")
	}
	return &productv1.GetProductResponse{Product: clone(product)}, nil
}

func (c *Catalog) ListProducts(_ context.Context, req *productv1.ListProductsRequest) (*productv1.ListProductsResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	response := &productv1.ListProductsResponse{}
	for _, product := range c.products {
		if req.GetStatus() != productv1.ProductStatus_PRODUCT_STATUS_UNSPECIFIED && product.GetStatus() != req.GetStatus() {
			continue
		}
		response.Products = append(response.Products, clone(product))
	}
	return response, nil
}

func (c *Catalog) UpdateProduct(_ context.Context, req *productv1.UpdateProductRequest) (*productv1.UpdateProductResponse, error) {
	if req == nil || req.GetProduct() == nil || req.GetProduct().GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "product ID is required")
	}
	if err := validate(req.GetProduct()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	current, ok := c.products[req.GetProduct().GetId()]
	if !ok {
		return nil, status.Error(codes.NotFound, "product not found")
	}
	updated := clone(req.GetProduct())
	updated.CreatedAt = current.CreatedAt
	updated.UpdatedAt = timestamppb.Now()
	c.products[updated.Id] = updated
	return &productv1.UpdateProductResponse{Product: clone(updated)}, nil
}

func (c *Catalog) ArchiveProduct(_ context.Context, req *productv1.ArchiveProductRequest) (*productv1.ArchiveProductResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	product, ok := c.products[req.GetId()]
	if !ok {
		return nil, status.Error(codes.NotFound, "product not found")
	}
	product.Status = productv1.ProductStatus_PRODUCT_STATUS_ARCHIVED
	product.UpdatedAt = timestamppb.Now()
	return &productv1.ArchiveProductResponse{Result: &emptypb.Empty{}}, nil
}

func validate(product *productv1.Product) error {
	if strings.TrimSpace(product.GetSku()) == "" || strings.TrimSpace(product.GetName()) == "" || strings.TrimSpace(product.GetCurrency()) == "" || product.GetPriceMinor() < 0 {
		return errInvalidProduct
	}
	return nil
}

func clone(product *productv1.Product) *productv1.Product {
	if product == nil {
		return nil
	}
	copy := *product
	return &copy
}
