package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewJWTService(t *testing.T) {
	service := NewJWTService("test-secret-key-32-chars-min!!", 24*time.Hour)

	if service == nil {
		t.Fatal("Expected service to be created")
	}
}

func TestJWTService_GenerateAndValidate(t *testing.T) {
	service := NewJWTService("test-secret-key-32-chars-minimum", 24*time.Hour)

	userID := uuid.New()
	companyID := uuid.New()
	email := "test@example.com"
	role := "OWNER"

	// Generate token
	token, err := service.GenerateToken(userID, companyID, email, role)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	if token == "" {
		t.Error("Token should not be empty")
	}

	// Validate token
	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != userID {
		t.Error("UserID mismatch")
	}
	if claims.CompanyID != companyID {
		t.Error("CompanyID mismatch")
	}
	if claims.Email != email {
		t.Error("Email mismatch")
	}
	if claims.Role != role {
		t.Error("Role mismatch")
	}
}

func TestJWTService_InvalidToken(t *testing.T) {
	service := NewJWTService("test-secret-key-32-chars-minimum", 24*time.Hour)

	// Invalid token
	_, err := service.ValidateToken("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}

	// Token signed with different secret
	otherService := NewJWTService("different-secret-key-32-chars-!!!", 24*time.Hour)
	token, _ := otherService.GenerateToken(uuid.New(), uuid.New(), "test@test.com", "USER")

	_, err = service.ValidateToken(token)
	if err == nil {
		t.Error("Expected error for token with wrong secret")
	}
}

func TestJWTService_RefreshToken(t *testing.T) {
	service := NewJWTService("test-secret-key-32-chars-minimum", 24*time.Hour)

	userID := uuid.New()
	companyID := uuid.New()

	// Generate original token
	token, _ := service.GenerateToken(userID, companyID, "test@example.com", "OWNER")
	claims, _ := service.ValidateToken(token)

	// Refresh token
	newToken, err := service.RefreshToken(claims)
	if err != nil {
		t.Fatalf("Failed to refresh token: %v", err)
	}
	if newToken == "" {
		t.Error("New token should not be empty")
	}

	// Validate new token
	newClaims, err := service.ValidateToken(newToken)
	if err != nil {
		t.Fatalf("Failed to validate refreshed token: %v", err)
	}

	if newClaims.UserID != userID {
		t.Error("UserID should match after refresh")
	}
}
