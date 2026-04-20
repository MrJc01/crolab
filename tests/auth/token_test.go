package auth_test

import (
	"strings"
	"testing"

	"github.com/crolab/core/internal/cloud"
)

func TestTokenGeneration(t *testing.T) {
	token := cloud.GenerateToken()

	if len(token) < 20 {
		t.Errorf("Token muito curto %d caracteres: %s", len(token), token)
	}

	token2 := cloud.GenerateToken()
	if token == token2 {
		t.Fatalf("Colisão de tokens críticos: ambou são %s", token)
	}

	if strings.ContainsAny(token, "=+&/\\") {
		t.Fatalf("Token com chars inseguros URL: %s", token)
	}
}
