package validator_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/better/backend/pkg/validator"
)

func TestValidator_ValidEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"test@example.com", true},
		{"user@domain.co", true},
		{"invalid", false},
		{"@missing.com", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			v := validator.New()
			v.ValidEmail(tt.email, "email")
			assert.Equal(t, tt.valid, v.Valid())
		})
	}
}

func TestValidator_ValidUsername(t *testing.T) {
	tests := []struct {
		username string
		valid    bool
	}{
		{"user123", true},
		{"test_user", true},
		{"ab", false},        // too short
		{"a", false},         // too short
		{"us er", false},     // space
		{"user@name", false}, // special char
	}

	for _, tt := range tests {
		t.Run(tt.username, func(t *testing.T) {
			v := validator.New()
			v.ValidUsername(tt.username, "username")
			assert.Equal(t, tt.valid, v.Valid())
		})
	}
}

func TestValidator_ValidPassword(t *testing.T) {
	tests := []struct {
		password string
		valid    bool
	}{
		{"password123", true},
		{"12345678", true},
		{"short", false},   // too short
		{"1234567", false}, // too short
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			v := validator.New()
			v.ValidPassword(tt.password, "password")
			assert.Equal(t, tt.valid, v.Valid())
		})
	}
}

func TestValidator_PositiveFloat(t *testing.T) {
	v := validator.New()
	v.PositiveFloat(1.5, "value")
	assert.True(t, v.Valid())

	v2 := validator.New()
	v2.PositiveFloat(-1.0, "value")
	assert.False(t, v2.Valid())

	v3 := validator.New()
	v3.PositiveFloat(0, "value")
	assert.False(t, v3.Valid())
}

func TestValidator_InList(t *testing.T) {
	v := validator.New()
	v.InList("BET", "type", []string{"DEPOSIT", "WITHDRAWAL", "BET", "WIN", "LOSS"})
	assert.True(t, v.Valid())

	v2 := validator.New()
	v2.InList("INVALID", "type", []string{"DEPOSIT", "WITHDRAWAL", "BET", "WIN", "LOSS"})
	assert.False(t, v2.Valid())
}
