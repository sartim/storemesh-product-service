package auth

import (
	"context"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryInterceptor(secret, issuer, audience string, oidc *OIDCValidator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		token, err := bearerToken(ctx)
		if err != nil {
			return nil, err
		}
		if oidc != nil {
			claims, err := oidc.ValidateClaims(token)
			if err != nil {
				return nil, status.Error(codes.Unauthenticated, "invalid OIDC bearer token")
			}
			if isProductMutation(info.FullMethod) && !hasAdminRole(claims) {
				return nil, status.Error(codes.PermissionDenied, "admin role is required")
			}
			return handler(ctx, req)
		}
		if secret == "" {
			return nil, status.Error(codes.Unauthenticated, "authentication is not configured")
		}
		claims := jwt.MapClaims{}
		parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, status.Error(codes.Unauthenticated, "unexpected signing method")
			}
			return []byte(secret), nil
		}, jwt.WithIssuer(issuer), jwt.WithAudience(audience))
		if err != nil || !parsed.Valid {
			return nil, status.Error(codes.Unauthenticated, "invalid bearer token")
		}
		return handler(ctx, req)
	}
}

func isProductMutation(method string) bool {
	switch method {
	case "/storemesh.product.v1.ProductCatalogService/CreateProduct",
		"/storemesh.product.v1.ProductCatalogService/UpdateProduct",
		"/storemesh.product.v1.ProductCatalogService/ArchiveProduct":
		return true
	default:
		return false
	}
}

func hasAdminRole(claims *OIDCClaims) bool {
	if claims == nil {
		return false
	}
	for _, role := range claims.RealmAccess.Roles {
		if strings.EqualFold(strings.TrimSpace(role), "admin") {
			return true
		}
	}
	return false
}

func bearerToken(ctx context.Context) (string, error) {
	values := metadata.ValueFromIncomingContext(ctx, "authorization")
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "authorization is required")
	}
	parts := strings.Fields(values[0])
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", status.Error(codes.Unauthenticated, "bearer authorization is required")
	}
	return parts[1], nil
}
