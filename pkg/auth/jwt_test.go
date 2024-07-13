package auth

import (
	"testing"

	"github.com/sfs/pkg/env"

	"github.com/alecthomas/assert/v2"
)

func TestTokenCreation(t *testing.T) {
	env.SetEnv(false)

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
	env.SetEnv(false)

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

func TestTokenExtraction(t *testing.T) {
	env.SetEnv(false)

	tok := NewT()
	userID := NewUUID()
	rawToken, err := tok.Create(userID)
	if err != nil {
		t.Fatalf("create token failed: %v", err)
	}
	fakeHeader := "Bearer " + rawToken
	tokenString, err := tok.Extract(fakeHeader)
	if err != nil {
		t.Fatalf("token extraction failed: %v", err)
	}
	assert.NotContains(t, tokenString, "Bearer")
	assert.NotEqual(t, "", tokenString)
}

func TestRetrieveSecret(t *testing.T) {
	env.SetEnv(false)
	s, err := GetSecret()
	if err != nil {
		t.Fail()
	}
	assert.NotEqual(t, nil, s)
	assert.NotEqual(t, 0, len(s))
}

func TestSecretGeneration(t *testing.T) {
	env.SetEnv(false)
	testSecret := GenSecret(64)

	assert.NotEqual(t, testSecret, "")
	assert.Equal(t, len(testSecret), 64)
}
