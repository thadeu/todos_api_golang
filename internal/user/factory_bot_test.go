package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestNewUser_WithCustomData_WithoutEncryptedPassword(t *testing.T) {
	user := NewUser[User](map[string]any{
		"Name":     "Test User",
		"Email":    "test@example.com",
		"Password": "12345678",
	})

	// Verifica se o EncryptedPassword foi gerado automaticamente
	assert.NotEmpty(t, user.EncryptedPassword)
	assert.Equal(t, "Test User", user.Name)
	assert.Equal(t, "test@example.com", user.Email)
}

func TestNewUser_WithCustomData_WithEncryptedPassword(t *testing.T) {
	customPassword := "custom12345678"
	encryptedPassword, _ := bcrypt.GenerateFromPassword([]byte(customPassword), bcrypt.DefaultCost)

	user := NewUser[User](map[string]any{
		"Name":              "Test User",
		"Email":             "test@example.com",
		"EncryptedPassword": string(encryptedPassword),
	})

	assert.Equal(t, string(encryptedPassword), user.EncryptedPassword)
	assert.Equal(t, "Test User", user.Name)
	assert.Equal(t, "test@example.com", user.Email)
}

func TestNewUser_WithMultipleMaps_EncryptedPasswordInSecondMap(t *testing.T) {
	customPassword := "custompassword123"

	user := NewUser[User](
		map[string]any{
			"Name":              "Test User",
			"Email":             "test@example.com",
			"EncryptedPassword": customPassword,
		},
	)

	assert.Equal(t, customPassword, user.EncryptedPassword)
	assert.Equal(t, "Test User", user.Name)
	assert.Equal(t, "test@example.com", user.Email)
}

func TestNewUser_WithMultipleMaps_NoEncryptedPassword(t *testing.T) {
	user := NewUser[User](
		map[string]any{
			"Name":  "Test User",
			"Email": "test@example.com",
		},
		map[string]any{
			"UUID": "some-uuid",
		},
	)

	assert.NotEmpty(t, user.EncryptedPassword)
	assert.Equal(t, "Test User", user.Name)
	assert.Equal(t, "test@example.com", user.Email)
}
