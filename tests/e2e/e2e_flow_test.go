package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/crolab/core/internal/cloud"
)

var testPort = 19900

func nextPort() string {
	testPort++
	return fmt.Sprintf(":%d", testPort)
}

// TestE2EFullFlow tests: admin creates plan → add pool → client registers → subscribes → runs job → checks billing
func TestE2EFullFlow(t *testing.T) {
	port := nextPort()
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "e2e.db")

	if err := cloud.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	cloud.SeedDefaultMachines()

	go cloud.StartCloudServer(port, "", "", "")
	time.Sleep(200 * time.Millisecond)

	base := "http://127.0.0.1" + port

	// 1. Admin registers (first user)
	_, adminData := postJSON(t, base+"/auth/register", map[string]interface{}{
		"email": "admin@e2e.ai", "password": "admin",
	})
	adminToken := adminData["token"].(string)
	if adminData["role"] != "admin" {
		t.Fatalf("primeiro user deveria ser admin")
	}

	// 2. Admin creates plan
	resp := authReqJ(t, "POST", base+"/admin/plans", adminToken, map[string]interface{}{
		"id": "start", "name": "Start", "vram": "6GB", "storage": "100GB",
		"price_hr": 0.30, "price_month": 29.90,
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create plan: %d", resp.StatusCode)
	}

	// 3. Admin adds pool entry
	resp = authReqJ(t, "POST", base+"/admin/pool/start", adminToken, map[string]interface{}{
		"priority": 1, "provider": "vastai", "label": "Vast T4", "address": "10.0.0.1:4422",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("add pool: %d", resp.StatusCode)
	}

	// 4. Client registers
	_, clientData := postJSON(t, base+"/auth/register", map[string]interface{}{
		"email": "dev@e2e.ai", "password": "dev123",
	})
	clientToken := clientData["token"].(string)
	if clientData["role"] != "client" {
		t.Fatalf("segundo user deveria ser client")
	}

	// 5. Client views plans
	resp = authReqJ(t, "GET", base+"/client/plans", clientToken, nil)
	if resp.StatusCode != 200 {
		t.Fatalf("client plans: %d", resp.StatusCode)
	}

	// 6. Client subscribes
	resp = authReqJ(t, "POST", base+"/client/subscribe", clientToken, map[string]interface{}{
		"plan_id": "start",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("subscribe: %d", resp.StatusCode)
	}

	// 7. Client checks subscription
	resp, subData := authReqJSON(t, "GET", base+"/client/subscription", clientToken, nil)
	plan := subData["plan"].(map[string]interface{})
	if plan["name"] != "Start" {
		t.Fatalf("plano errado: %v", plan["name"])
	}

	// 8. Client connects personal machine
	resp = authReqJ(t, "POST", base+"/client/machines", clientToken, map[string]interface{}{
		"name": "minha-gpu", "address": "192.168.1.10:4422", "provider": "personal",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("connect machine: %d", resp.StatusCode)
	}

	// 9. Client runs job
	resp, jobData := authReqJSON(t, "POST", base+"/client/run", clientToken, map[string]interface{}{
		"plan_id": "start",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("run job: %d", resp.StatusCode)
	}
	jobID := jobData["job_id"].(string)
	if jobID == "" {
		t.Fatal("job_id vazio")
	}

	// 10. Client checks jobs
	resp, _ = authReqJSON(t, "GET", base+"/client/jobs", clientToken, nil)
	if resp.StatusCode != 200 {
		t.Fatalf("jobs list: %d", resp.StatusCode)
	}

	// 11. Client buys credits
	resp = authReqJ(t, "POST", base+"/billing/purchase", clientToken, map[string]interface{}{
		"amount": 50.0,
	})
	if resp.StatusCode != 200 {
		t.Fatalf("purchase: %d", resp.StatusCode)
	}

	// 12. Check billing
	_, meData := authReqJSON(t, "GET", base+"/auth/me", clientToken, nil)
	credits := meData["credits"].(float64)
	if credits != 59.70 {
		t.Fatalf("créditos esperados 59.70, obteve %.2f", credits)
	}

	// 13. Admin checks dashboard
	_, dashData := authReqJSON(t, "GET", base+"/admin/dashboard", adminToken, nil)
	if dashData["users_total"].(float64) != 2 {
		t.Errorf("dashboard users: %v", dashData["users_total"])
	}

	// 14. Admin checks logs
	resp = authReqJ(t, "GET", base+"/admin/logs", adminToken, nil)
	if resp.StatusCode != 200 {
		t.Fatalf("logs: %d", resp.StatusCode)
	}

	// 15. Client unsubscribes
	resp = authReqJ(t, "DELETE", base+"/client/subscription", clientToken, nil)
	if resp.StatusCode != 200 {
		t.Fatalf("unsubscribe: %d", resp.StatusCode)
	}

	t.Log("✅ E2E COMPLETO: register → plan → pool → subscribe → connect → run → billing → dashboard → unsubscribe")
}

// --- Helpers ---

func postJSON(t *testing.T, url string, body map[string]interface{}) (*http.Response, map[string]interface{}) {
	t.Helper()
	b, _ := json.Marshal(body)
	resp, err := http.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	data, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return resp, result
}

func authReqJ(t *testing.T, method, url, token string, body interface{}) *http.Response {
	t.Helper()
	resp, _ := authReqJSON(t, method, url, token, body)
	return resp
}

func authReqJSON(t *testing.T, method, url, token string, body interface{}) (*http.Response, map[string]interface{}) {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, url, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	data, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return resp, result
}

var _ = os.TempDir
