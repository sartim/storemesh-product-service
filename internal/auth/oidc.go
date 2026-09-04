package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type OIDCValidator struct {
	issuer, audience, jwksURL string
	mu                        sync.RWMutex
	keys                      map[string]*rsa.PublicKey
}

type oidcDiscovery struct {
	JWKSURL string `json:"jwks_uri"`
}
type oidcJWKS struct {
	Keys []struct{ Kid, Kty, N, E string } `json:"keys"`
}

func NewOIDCValidator(issuer, audience string) (*OIDCValidator, error) {
	issuer = strings.TrimRight(strings.TrimSpace(issuer), "/")
	if issuer == "" || audience == "" {
		return nil, nil
	}
	v := &OIDCValidator{issuer: issuer, audience: audience}
	var discovery oidcDiscovery
	if err := getJSON(issuer+"/.well-known/openid-configuration", &discovery); err != nil {
		return nil, err
	}
	v.jwksURL = discovery.JWKSURL
	if err := v.refresh(); err != nil {
		return nil, err
	}
	return v, nil
}

func (v *OIDCValidator) refresh() error {
	var set oidcJWKS
	if err := getJSON(v.jwksURL, &set); err != nil {
		return err
	}
	keys := map[string]*rsa.PublicKey{}
	for _, k := range set.Keys {
		if k.Kty != "RSA" {
			continue
		}
		n, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			return err
		}
		e, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			return err
		}
		exponent := 0
		for _, b := range e {
			exponent = exponent*256 + int(b)
		}
		keys[k.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: exponent}
	}
	v.mu.Lock()
	v.keys = keys
	v.mu.Unlock()
	return nil
}

func (v *OIDCValidator) Validate(raw string) error {
	token, err := jwt.ParseWithClaims(raw, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodRS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		kid, _ := token.Header["kid"].(string)
		v.mu.RLock()
		key := v.keys[kid]
		v.mu.RUnlock()
		if key == nil {
			return nil, fmt.Errorf("unknown signing key")
		}
		return key, nil
	}, jwt.WithIssuer(v.issuer), jwt.WithAudience(v.audience), jwt.WithLeeway(30*time.Second))
	if err != nil || !token.Valid {
		return fmt.Errorf("invalid OIDC token: %w", err)
	}
	return nil
}

func getJSON(url string, target any) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode/100 != 2 {
		return fmt.Errorf("OIDC endpoint returned %s", response.Status)
	}
	return json.NewDecoder(response.Body).Decode(target)
}
