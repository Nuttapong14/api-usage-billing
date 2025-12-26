package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrMissingAuthHeader = errors.New("missing authorization header")
	ErrInvalidAuthFormat = errors.New("invalid authorization header format")
	ErrInvalidToken      = errors.New("invalid token")
	ErrTokenExpired      = errors.New("token has expired")
	ErrInvalidIssuer     = errors.New("invalid token issuer")
)

// KeycloakConfig holds Keycloak configuration
type KeycloakConfig struct {
	URL          string
	Realm        string
	ClientID     string
	ClientSecret string
}

// KeycloakClaims represents JWT claims from Keycloak
type KeycloakClaims struct {
	jwt.RegisteredClaims
	Email         string                 `json:"email"`
	EmailVerified bool                   `json:"email_verified"`
	Name          string                 `json:"name"`
	PreferredUser string                 `json:"preferred_username"`
	RealmAccess   RealmAccess            `json:"realm_access"`
	ResourceAccess map[string]RoleAccess `json:"resource_access"`
	Scope         string                 `json:"scope"`
}

// RealmAccess contains realm-level roles
type RealmAccess struct {
	Roles []string `json:"roles"`
}

// RoleAccess contains client-level roles
type RoleAccess struct {
	Roles []string `json:"roles"`
}

// JWKS represents a JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
	Alg string `json:"alg"`
}

// KeycloakAuthMiddleware handles JWT validation with Keycloak
type KeycloakAuthMiddleware struct {
	config    KeycloakConfig
	keys      map[string]*rsa.PublicKey
	keysMutex sync.RWMutex
	lastFetch time.Time
}

// NewKeycloakAuthMiddleware creates a new Keycloak auth middleware
func NewKeycloakAuthMiddleware(config KeycloakConfig) *KeycloakAuthMiddleware {
	return &KeycloakAuthMiddleware{
		config: config,
		keys:   make(map[string]*rsa.PublicKey),
	}
}

// Handler returns the Fiber middleware handler
func (m *KeycloakAuthMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": ErrMissingAuthHeader.Error(),
			})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": ErrInvalidAuthFormat.Error(),
			})
		}

		tokenString := parts[1]

		// Parse and validate token
		claims, err := m.validateToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": err.Error(),
			})
		}

		// Store claims in context
		c.Locals("user_id", claims.Subject)
		c.Locals("email", claims.Email)
		c.Locals("name", claims.Name)
		c.Locals("claims", claims)

		// Parse user UUID if valid
		if userID, err := uuid.Parse(claims.Subject); err == nil {
			c.Locals("user_uuid", userID)
		}

		return c.Next()
	}
}

func (m *KeycloakAuthMiddleware) validateToken(tokenString string) (*KeycloakClaims, error) {
	// Parse token without validation first to get kid
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &KeycloakClaims{})
	if err != nil {
		return nil, ErrInvalidToken
	}

	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Get public key for this kid
	publicKey, err := m.getPublicKey(kid)
	if err != nil {
		return nil, err
	}

	// Parse and validate token with public key
	token, err = jwt.ParseWithClaims(tokenString, &KeycloakClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*KeycloakClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Validate issuer
	expectedIssuer := fmt.Sprintf("%s/realms/%s", m.config.URL, m.config.Realm)
	if claims.Issuer != expectedIssuer {
		return nil, ErrInvalidIssuer
	}

	return claims, nil
}

func (m *KeycloakAuthMiddleware) getPublicKey(kid string) (*rsa.PublicKey, error) {
	m.keysMutex.RLock()
	key, exists := m.keys[kid]
	m.keysMutex.RUnlock()

	if exists {
		return key, nil
	}

	// Fetch JWKS if key not found or cache expired
	if err := m.fetchJWKS(); err != nil {
		return nil, err
	}

	m.keysMutex.RLock()
	key, exists = m.keys[kid]
	m.keysMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("key with kid %s not found", kid)
	}

	return key, nil
}

func (m *KeycloakAuthMiddleware) fetchJWKS() error {
	m.keysMutex.Lock()
	defer m.keysMutex.Unlock()

	// Skip if recently fetched
	if time.Since(m.lastFetch) < time.Minute {
		return nil
	}

	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", m.config.URL, m.config.Realm)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, jwksURL, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch JWKS: status %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return err
	}

	// Parse and store keys
	for _, jwk := range jwks.Keys {
		if jwk.Kty != "RSA" || jwk.Use != "sig" {
			continue
		}

		publicKey, err := parseRSAPublicKey(jwk)
		if err != nil {
			continue
		}

		m.keys[jwk.Kid] = publicKey
	}

	m.lastFetch = time.Now()
	return nil
}

func parseRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, err
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nBytes)
	e := int(new(big.Int).SetBytes(eBytes).Int64())

	return &rsa.PublicKey{N: n, E: e}, nil
}

// HasRole checks if the user has a specific realm role
func HasRole(c *fiber.Ctx, role string) bool {
	claims, ok := c.Locals("claims").(*KeycloakClaims)
	if !ok {
		return false
	}

	for _, r := range claims.RealmAccess.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasClientRole checks if the user has a specific client role
func HasClientRole(c *fiber.Ctx, clientID, role string) bool {
	claims, ok := c.Locals("claims").(*KeycloakClaims)
	if !ok {
		return false
	}

	access, ok := claims.ResourceAccess[clientID]
	if !ok {
		return false
	}

	for _, r := range access.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// RequireRole creates middleware that requires a specific role
func RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !HasRole(c, role) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "forbidden",
				"message": "insufficient permissions",
			})
		}
		return c.Next()
	}
}
