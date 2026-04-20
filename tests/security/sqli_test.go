package security_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/crolab/core/internal/cloud"
)

func TestSQLInjectionLogin(t *testing.T) {
	// 1. Setup isolated memory DB equivalent to InitDB without touching filesystem if possible, or temp file.
	dbPath := t.TempDir() + "/sqli_test.db"
	if err := cloud.InitDB(dbPath); err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}
	defer os.Remove(dbPath)

	// Admin mock
	cloud.DBCreateUser("admin@crolab.com", "secure123", "admin", "127.0.0.1")

	// 2. Instantiate mux that handles /auth/login
	mux := cloud.BuildMux("./web")

	// 3. Attack Payloads
	payloads := []string{
		"admin@crolab.com' OR '1'='1",
		"admin@crolab.com\" OR \"1\"=\"1",
		"admin@crolab.com'; DROP TABLE users;--",
		"' OR 1=1;--",
	}

	for _, payload := range payloads {
		t.Run("Payload: "+payload, func(t *testing.T) {
			body := map[string]string{
				"email":    payload,
				"password": "wrongpassword",
			}
			b, _ := json.Marshal(body)

			req, _ := http.NewRequest("POST", "/auth/login", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			// Expecting 401 Unauthorized for bad creds vs 200 OK which means a successful bypass
			if rr.Code == http.StatusOK {
				t.Fatalf(" CRITICAL VULNERABILITY: SQL Injection bypassed login with payload: %s", payload)
			}
			if !strings.Contains(rr.Body.String(), "credenciais") {
				t.Logf("Safe failure for payload, response: %v", rr.Body.String())
			}
		})
	}
}
