package security_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/crolab/core/internal/cloud"
)

func TestBruteForceRateLimit(t *testing.T) {
	dbPath := t.TempDir() + "/bruteforce_test.db"
	if err := cloud.InitDB(dbPath); err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}
	defer os.Remove(dbPath)

	mux := cloud.BuildMux("./web")

	requests := 100
	var wg sync.WaitGroup
	wg.Add(requests)

	body := map[string]string{"email": "spam@app.com", "password": "123"}
	b, _ := json.Marshal(body)

	start := time.Now()
	tooManyRequestsCount := 0

	var mu sync.Mutex

	for i := 0; i < requests; i++ {
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest("POST", "/auth/login", bytes.NewReader(b))
			// Simulate same IP attacking
			req.RemoteAddr = "10.0.0.5:12345"
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			mu.Lock()
			if rr.Code == 429 { // Too Many Requests expected from RateLimit Middleware
				tooManyRequestsCount++
			}
			mu.Unlock()
		}()
	}

	wg.Wait()
	duration := time.Since(start)

	t.Logf("Fired %d requests in %v. Blocked requests: %d", requests, duration, tooManyRequestsCount)

	if tooManyRequestsCount == 0 {
		t.Fatalf("CRITICAL: Rate limit middleware failed or missing! All %d requests bypassed and likely consumed heavy Bcrypt CPU hashing.", requests)
	}
}
