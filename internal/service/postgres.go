package service

import (
	"context"
	"database/sql"
	"strings"

	"github.com/google/uuid"
	productv1 "github.com/sartim/storemesh-product-service/gen/storemesh/product/v1"
	"github.com/sartim/storemesh-product-service/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PersistentCatalog struct {
	productv1.UnimplementedProductCatalogServiceServer
	store *repository.ProductStore
}

func NewPersistentCatalog(store *repository.ProductStore) *PersistentCatalog {
	return &PersistentCatalog{store: store}
}

func (c *PersistentCatalog) CreateProduct(ctx context.Context, req *productv1.CreateProductRequest) (*productv1.CreateProductResponse, error) {
	if req == nil || req.GetProduct() == nil || validate(req.GetProduct()) != nil {
		return nil, status.Error(codes.InvalidArgument, errInvalidProduct.Error())
	}
	p := clone(req.GetProduct())
	p.Id, p.Status = uuid.NewString(), productv1.ProductStatus_PRODUCT_STATUS_ACTIVE
	p.CreatedAt, p.UpdatedAt = timestamppb.Now(), timestamppb.Now()
	if err := c.store.Insert(ctx, p); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, status.Errorf(codes.AlreadyExists, "product SKU %q already exists", p.Sku)
		}
		return nil, status.Error(codes.Internal, "create product")
	}
	return &productv1.CreateProductResponse{Product: p}, nil
}

func (c *PersistentCatalog) GetProduct(ctx context.Context, req *productv1.GetProductRequest) (*productv1.GetProductResponse, error) {
	p, err := c.store.Find(ctx, req.GetId())
	if err == sql.ErrNoRows {
		return nil, status.Error(codes.NotFound, "product not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "get product")
	}
	return &productv1.GetProductResponse{Product: p}, nil
}

func (c *PersistentCatalog) ListProducts(ctx context.Context, req *productv1.ListProductsRequest) (*productv1.ListProductsResponse, error) {
	pageSize, offset, err := pageParameters(req)
	if err != nil {
		return nil, err
	}
	products, hasMore, err := c.store.List(ctx, req.GetStatus(), pageSize, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, "list products")
	}
	response := &productv1.ListProductsResponse{Products: products}
	if hasMore {
		response.NextPageToken = nextPageToken(offset, pageSize, offset+len(products)+1)
	}
	return response, nil
}

func (c *PersistentCatalog) UpdateProduct(ctx context.Context, req *productv1.UpdateProductRequest) (*productv1.UpdateProductResponse, error) {
	if req == nil || req.GetProduct() == nil || req.GetProduct().GetId() == "" || validate(req.GetProduct()) != nil {
		return nil, status.Error(codes.InvalidArgument, errInvalidProduct.Error())
	}
	p := clone(req.GetProduct())
	current, err := c.store.Find(ctx, p.Id)
	if err == sql.ErrNoRows {
		return nil, status.Error(codes.NotFound, "product not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "get product")
	}
	p.CreatedAt, p.UpdatedAt = current.CreatedAt, timestamppb.Now()
	if err := c.store.Update(ctx, p); err != nil {
		return nil, status.Error(codes.Internal, "update product")
	}
	return &productv1.UpdateProductResponse{Product: p}, nil
}

func (c *PersistentCatalog) ArchiveProduct(ctx context.Context, req *productv1.ArchiveProductRequest) (*productv1.ArchiveProductResponse, error) {
	if _, err := c.store.Find(ctx, req.GetId()); err == sql.ErrNoRows {
		return nil, status.Error(codes.NotFound, "product not found")
	} else if err != nil {
		return nil, status.Error(codes.Internal, "get product")
	}
	if err := c.store.Archive(ctx, req.GetId(), productv1.ProductStatus_PRODUCT_STATUS_ARCHIVED, timestamppb.Now().AsTime()); err != nil {
		return nil, status.Error(codes.Internal, "archive product")
	}
	return &productv1.ArchiveProductResponse{Result: &emptypb.Empty{}}, nil
}
