package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestOIDCValidatorRefreshesJWKSForRotatedKey(t *testing.T) {
	initial, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	rotated, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	var mu sync.RWMutex
	activeKey, activeKID := &initial.PublicKey, "initial"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.RLock()
		defer mu.RUnlock()
		if r.URL.Path == "/.well-known/openid-configuration" {
			_ = json.NewEncoder(w).Encode(map[string]string{"issuer": "http://" + r.Host, "jwks_uri": "http://" + r.Host + "/keys"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{"kid": activeKID, "kty": "RSA", "n": base64.RawURLEncoding.EncodeToString(activeKey.N.Bytes()), "e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(activeKey.E)).Bytes())}}})
	}))
	defer server.Close()
	validator, err := NewOIDCValidator(server.URL, "storemesh-bff")
	if err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	activeKey, activeKID = &rotated.PublicKey, "rotated"
	mu.Unlock()
	claims := &OIDCClaims{RegisteredClaims: jwt.RegisteredClaims{Issuer: server.URL, Subject: "customer-1", Audience: jwt.ClaimStrings{"storemesh-bff"}, ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))}}
	claims.RealmAccess.Roles = []string{"customer"}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "rotated"
	raw, err := token.SignedString(rotated)
	if err != nil {
		t.Fatal(err)
	}
	validatedClaims, err := validator.ValidateClaims(raw)
	if err != nil {
		t.Fatalf("validator did not accept rotated signing key: %v", err)
	}
	if validatedClaims.Subject != "customer-1" || len(validatedClaims.RealmAccess.Roles) != 1 || validatedClaims.RealmAccess.Roles[0] != "customer" {
		t.Fatalf("validated claims did not preserve identity and roles: %#v", validatedClaims)
	}

	interceptor := UnaryInterceptor("", "", "", validator)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+raw))
	called := false
	_, err = interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/storemesh.product.v1.ProductCatalogService/CreateProduct"}, func(context.Context, any) (any, error) {
		called = true
		return nil, nil
	})
	if status.Code(err) != codes.PermissionDenied || called {
		t.Fatalf("customer write should be denied before handler; called=%v err=%v", called, err)
	}
	_, err = interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/storemesh.product.v1.ProductCatalogService/ListProducts"}, func(context.Context, any) (any, error) {
		called = true
		return nil, nil
	})
	if err != nil || !called {
		t.Fatalf("authenticated customer read should be allowed; called=%v err=%v", called, err)
	}
}
