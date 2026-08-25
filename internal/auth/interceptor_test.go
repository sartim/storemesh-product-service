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
	interceptor := UnaryInterceptor("secret", "issuer", "audience")
	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/test"}, func(context.Context, any) (any, error) { return "ok", nil })
	if status.Code(err) != codes.Unauthenticated { t.Fatalf("expected unauthenticated, got %v", err) }

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Basic token"))
	_, err = interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test"}, func(context.Context, any) (any, error) { return "ok", nil })
	if status.Code(err) != codes.Unauthenticated { t.Fatalf("expected unauthenticated, got %v", err) }
}
