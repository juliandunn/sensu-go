package jwt

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	utilbytes "github.com/sensu/sensu-go/util/bytes"
)

// Define the key type to avoid key collisions in context
type key int

const (
	// claimsKey contains the key name used to store the JWT claims within
	// the context of a request
	claimsKey key = iota
)

var (
	defaultExpiration = time.Minute * time.Duration(15)
	secret            []byte
)

// AccessToken creates a new access token and returns it in both JWT and
// signed format, along with any error
func AccessToken(username string) (*jwt.Token, string, error) {
	// Create a unique identifier for the token
	jti, err := utilbytes.Random(16)
	if err != nil {
		return nil, "", err
	}

	claims := types.Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(defaultExpiration).Unix(),
			Id:        hex.EncodeToString(jti),
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)

	// Sign the token as a string using the secret
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return nil, "", err
	}

	return token, tokenString, nil
}

// GetClaims returns the claims from a token
func GetClaims(token *jwt.Token) (*types.Claims, error) {
	if claims, ok := token.Claims.(*types.Claims); ok {
		return claims, nil
	}

	return nil, fmt.Errorf("Could not parse the token claims")
}

// GetClaimsFromContext retrieves the JWT claims from the request context
func GetClaimsFromContext(r *http.Request) *types.Claims {
	if value := r.Context().Value(claimsKey); value != nil {
		claims, ok := value.(*types.Claims)
		if !ok {
			return nil
		}
		return claims
	}
	return nil
}

// ExtractBearerToken retrieves the bearer token from a request and returns the
// JWT
func ExtractBearerToken(r *http.Request) string {
	// Does a bearer token was provided in the Authorization header?
	var tokenString string
	tokens, ok := r.Header["Authorization"]
	if ok && len(tokens) >= 1 {
		tokenString = tokens[0]
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	}

	return tokenString
}

// InitSecret initializes and retrieves the secret for our signing tokens
func InitSecret(store store.Store) error {
	var s []byte
	var err error

	// Retrieve the secret
	if secret == nil {
		s, err = store.GetJWTSecret()
		if err != nil {
			// The secret does not exist, we need to create one
			s, err = utilbytes.Random(32)
			if err != nil {
				return err
			}

			// Add the secret to the store
			err = store.CreateJWTSecret(s)
			if err != nil {
				return err
			}
		}

		// Set the secret so it's available accross the package
		secret = s
	}

	return nil
}

// parseToken takes a signed token and parse it to verify its integrity
func parseToken(tokenString string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, &types.Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// secret is a []byte containing the secret
		return secret, nil
	})
}

// RefreshToken returns a refresh token for a specific user
func RefreshToken(username string) (string, error) {
	// Create a unique identifier for the token
	jti, err := utilbytes.Random(16)
	if err != nil {
		return "", err
	}

	claims := &jwt.StandardClaims{
		Id:      hex.EncodeToString(jti),
		Subject: username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token as a string using the secret
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// SetClaimsIntoContext adds the token claims into the request context for
// easier consumption later
func SetClaimsIntoContext(r *http.Request, token *jwt.Token) context.Context {
	claims, _ := token.Claims.(*types.Claims)
	return context.WithValue(r.Context(), claimsKey, claims)
}

// ValidateExpiredToken verifies that the provided token is valid, even if
// it's expired
func ValidateExpiredToken(tokenString string) (*jwt.Token, error) {
	token, err := parseToken(tokenString)
	if token == nil {
		return nil, err
	}

	if _, ok := token.Claims.(*types.Claims); ok {
		if token.Valid {
			return token, nil
		}

		// Inspect the error to determine the cause
		if validationError, ok := err.(*jwt.ValidationError); ok {
			if validationError.Errors&jwt.ValidationErrorExpired != 0 {
				// We already know that the token is expired and we don't care at that
				// point, we simply want to know if there's any other error
				validationError.Errors ^= jwt.ValidationErrorExpired
			}

			// Return the token if we have no other validation error
			if validationError.Errors == 0 {
				return token, nil
			}
		}
	}

	return nil, err
}

// ValidateToken verifies that the provided token is valid
func ValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := parseToken(tokenString)
	if token == nil {
		return nil, err
	}

	if _, ok := token.Claims.(*types.Claims); ok && token.Valid {
		return token, nil
	}

	return nil, err
}
