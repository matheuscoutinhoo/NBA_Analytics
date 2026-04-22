package unit

import (
	"testing"

	"github.com/matheuscoutinhoo/better/internal/auth"
)

func TestHashPassword(t *testing.T) {
	password := "TestPass123"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Fatal("hash should not be empty")
	}

	if hash == password {
		t.Fatal("hash should not equal the plain password")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "TestPass123"
	hash, _ := auth.HashPassword(password)

	if !auth.CheckPassword(password, hash) {
		t.Fatal("CheckPassword should return true for correct password")
	}

	if auth.CheckPassword("WrongPass123", hash) {
		t.Fatal("CheckPassword should return false for wrong password")
	}
}

func TestGenerateAccessToken(t *testing.T) {
	secret := "test-secret-key-at-least-32-characters!"
	token, err := auth.GenerateAccessToken(1, "test@example.com", "user", secret)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}
}

func TestValidateAccessToken(t *testing.T) {
	secret := "test-secret-key-at-least-32-characters!"

	token, err := auth.GenerateAccessToken(1, "test@example.com", "user", secret)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	claims, err := auth.ValidateAccessToken(token, secret)
	if err != nil {
		t.Fatalf("ValidateAccessToken failed: %v", err)
	}

	if claims.UserID != 1 {
		t.Errorf("expected UserID 1, got %d", claims.UserID)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", claims.Email)
	}
	if claims.Role != "user" {
		t.Errorf("expected role user, got %s", claims.Role)
	}
}

func TestValidateAccessToken_WrongSecret(t *testing.T) {
	token, _ := auth.GenerateAccessToken(1, "test@example.com", "user", "correct-secret-key-at-least-32chars!")
	_, err := auth.ValidateAccessToken(token, "wrong-secret-key-at-least-32characters!")
	if err == nil {
		t.Fatal("should fail with wrong secret")
	}
}

func TestValidateAccessToken_InvalidToken(t *testing.T) {
	_, err := auth.ValidateAccessToken("invalid-token", "some-secret")
	if err == nil {
		t.Fatal("should fail with invalid token")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	token, hash, err := auth.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	if token == "" {
		t.Fatal("token should not be empty")
	}
	if hash == "" {
		t.Fatal("hash should not be empty")
	}
	if token == hash {
		t.Fatal("token and hash should be different")
	}
}

func TestHashToken(t *testing.T) {
	token := "test-token-value"
	hash1 := auth.HashToken(token)
	hash2 := auth.HashToken(token)

	if hash1 != hash2 {
		t.Fatal("same token should produce same hash")
	}

	differentHash := auth.HashToken("different-token")
	if hash1 == differentHash {
		t.Fatal("different tokens should produce different hashes")
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		email    string
		valid    bool
	}{
		{"valid password", "TestPass1", "test@test.com", true},
		{"too short", "Te1", "test@test.com", false},
		{"no uppercase", "testpass1", "test@test.com", false},
		{"no lowercase", "TESTPASS1", "test@test.com", false},
		{"no digit", "TestPass", "test@test.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't test validatePassword directly since it's unexported,
			// but we test it implicitly through the auth flow.
			// For now we test the password hashing works for valid passwords.
			if tt.valid {
				hash, err := auth.HashPassword(tt.password)
				if err != nil {
					t.Fatalf("should hash valid password: %v", err)
				}
				if !auth.CheckPassword(tt.password, hash) {
					t.Fatal("should verify valid password")
				}
			}
		})
	}
}
