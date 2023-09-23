package auth

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestTokenCreation(t *testing.T) {
	tok := NewT()
	userID := NewUUID()
	tokenString, err := tok.Create(userID)
	if err != nil {
		t.Fatalf("create token failed: %v", err)
	}
	assert.NotEqual(t, "", tokenString)
	assert.Equal(t, tok.Jwt, tokenString)
}

func TestTokenVerification(t *testing.T) {
	tok := NewT()
	userID := NewUUID()
	tokenString, err := tok.Create(userID)
	if err != nil {
		t.Fatalf("create token failed: %v", err)
	}
	id, err := tok.Verify(tokenString)
	if err != nil {
		t.Fatalf("token validation failed: %v", err)
	}
	assert.NotEqual(t, "", id)
	assert.Equal(t, userID, id)
}
