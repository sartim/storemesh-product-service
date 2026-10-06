package auth

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestUnaryInterceptorRequiresBearerToken(t *testing.T) {
	interceptor := UnaryInterceptor("secret", "issuer", "audience", nil)
	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/test"}, func(context.Context, any) (any, error) { return "ok", nil })
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected unauthenticated, got %v", err)
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Basic token"))
	_, err = interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test"}, func(context.Context, any) (any, error) { return "ok", nil })
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
}

func TestProductMutationRequiresAdminRole(t *testing.T) {
	if !isProductMutation("/storemesh.product.v1.ProductCatalogService/CreateProduct") ||
		!isProductMutation("/storemesh.product.v1.ProductCatalogService/UpdateProduct") ||
		!isProductMutation("/storemesh.product.v1.ProductCatalogService/ArchiveProduct") {
		t.Fatal("catalog mutation method was not identified")
	}
	if isProductMutation("/storemesh.product.v1.ProductCatalogService/ListProducts") {
		t.Fatal("catalog read method must not be treated as a mutation")
	}
	if hasAdminRole(&OIDCClaims{}) {
		t.Fatal("customer without roles must not be an admin")
	}
	admin := &OIDCClaims{}
	admin.RealmAccess.Roles = []string{"customer", "admin"}
	if !hasAdminRole(admin) {
		t.Fatal("realm admin role should authorize catalog mutations")
	}
}
