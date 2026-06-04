package service

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/modulops/auth-service/internal/config"
	"github.com/modulops/auth-service/internal/models"
)

// buildTestService crée un AuthService minimal pour tester la logique JWT (sans DB).
func buildTestService(secret string, expiration time.Duration) *AuthService {
	return &AuthService{
		repo: nil, // non utilisé dans les tests JWT
		cfg: &config.Config{
			JWTSecret:     secret,
			JWTExpiration: expiration,
			JWTIssuer:     "kura-test",
		},
	}
}

// makeToken génère un token JWT signé avec les paramètres fournis.
func makeToken(secret, userID, email string, roles []string, expiry time.Time) string {
	claims := &config.Claims{
		UserID: userID,
		Email:  email,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "kura-test",
			Subject:   userID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(secret))
	return signed
}

func TestValidateToken_ValidToken(t *testing.T) {
	svc := buildTestService("test-secret-key", time.Hour)
	token := makeToken("test-secret-key", "user-123", "alice@example.com", []string{"user"}, time.Now().Add(time.Hour))

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() erreur inattendue : %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("UserID = %q, attendu %q", claims.UserID, "user-123")
	}
	if claims.Email != "alice@example.com" {
		t.Errorf("Email = %q, attendu %q", claims.Email, "alice@example.com")
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	svc := buildTestService("test-secret-key", time.Hour)
	token := makeToken("test-secret-key", "user-123", "alice@example.com", []string{"user"}, time.Now().Add(-time.Hour))

	_, err := svc.ValidateToken(token)
	if err == nil {
		t.Fatal("ValidateToken() devrait retourner une erreur pour un token expiré")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	svc := buildTestService("secret-correct", time.Hour)
	token := makeToken("secret-wrong", "user-123", "alice@example.com", []string{"user"}, time.Now().Add(time.Hour))

	_, err := svc.ValidateToken(token)
	if err == nil {
		t.Fatal("ValidateToken() devrait rejeter un token signé avec un mauvais secret")
	}
}

func TestValidateToken_TamperedToken(t *testing.T) {
	svc := buildTestService("test-secret-key", time.Hour)
	token := makeToken("test-secret-key", "user-123", "alice@example.com", []string{"user"}, time.Now().Add(time.Hour))

	// Altération d'un caractère dans la signature
	tampered := token[:len(token)-4] + "XXXX"
	_, err := svc.ValidateToken(tampered)
	if err == nil {
		t.Fatal("ValidateToken() devrait rejeter un token altéré")
	}
}

func TestValidateToken_EmptyToken(t *testing.T) {
	svc := buildTestService("test-secret-key", time.Hour)

	_, err := svc.ValidateToken("")
	if err == nil {
		t.Fatal("ValidateToken() devrait retourner une erreur pour un token vide")
	}
}

func TestValidateToken_AdminRolePreserved(t *testing.T) {
	svc := buildTestService("test-secret-key", time.Hour)
	token := makeToken("test-secret-key", "admin-456", "admin@kura.io", []string{"admin", "user"}, time.Now().Add(time.Hour))

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() erreur inattendue : %v", err)
	}
	if len(claims.Roles) != 2 {
		t.Errorf("Roles = %v, attendu 2 rôles", claims.Roles)
	}
}

func TestRegister_RejectsEmptyFields(t *testing.T) {
	svc := buildTestService("test-secret-key", time.Hour)

	cases := []struct {
		name string
		req  *models.RegisterRequest
	}{
		{"email vide", &models.RegisterRequest{Email: "", Password: "pass123", Username: "alice"}},
		{"password vide", &models.RegisterRequest{Email: "alice@example.com", Password: "", Username: "alice"}},
		{"username vide", &models.RegisterRequest{Email: "alice@example.com", Password: "pass123", Username: ""}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Register(tc.req)
			if err == nil {
				t.Fatalf("Register() devrait retourner une erreur pour : %s", tc.name)
			}
		})
	}
}
