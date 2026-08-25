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

func UnaryInterceptor(secret, issuer, audience string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		token, err := bearerToken(ctx)
		if err != nil { return nil, err }
		claims := jwt.MapClaims{}
		parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok { return nil, status.Error(codes.Unauthenticated, "unexpected signing method") }
			return []byte(secret), nil
		}, jwt.WithIssuer(issuer), jwt.WithAudience(audience))
		if err != nil || !parsed.Valid { return nil, status.Error(codes.Unauthenticated, "invalid bearer token") }
		return handler(ctx, req)
	}
}

func bearerToken(ctx context.Context) (string, error) {
	values := metadata.ValueFromIncomingContext(ctx, "authorization")
	if len(values) == 0 { return "", status.Error(codes.Unauthenticated, "authorization is required") }
	parts := strings.Fields(values[0])
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") { return "", status.Error(codes.Unauthenticated, "bearer authorization is required") }
	return parts[1], nil
}
