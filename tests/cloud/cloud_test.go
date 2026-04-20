package cloud_test

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

var testPort int = 18800

func nextPort() string {
	testPort++
	return fmt.Sprintf(":%d", testPort)
}

func startTestServer(t *testing.T) string {
	t.Helper()
	port := nextPort()

	// Each test gets its own temp DB
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "test.db")
	if err := cloud.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	cloud.SeedDefaultMachines()

	go cloud.StartCloudServer(port, "", "", "")
	time.Sleep(150 * time.Millisecond)
	return "http://127.0.0.1" + port
}

func post(url string, body map[string]interface{}) (*http.Response, map[string]interface{}) {
	b, _ := json.Marshal(body)
	resp, err := http.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return &http.Response{StatusCode: 503}, map[string]interface{}{"error": err.Error()}
	}
	data, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return resp, result
}

func get(url string, token string) (*http.Response, map[string]interface{}) {
	req, _ := http.NewRequest("GET", url, nil)
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &http.Response{StatusCode: 503}, map[string]interface{}{"error": err.Error()}
	}
	data, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return resp, result
}

func authReq(method, url, token string, body interface{}) (*http.Response, map[string]interface{}) {
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
		return &http.Response{StatusCode: 503}, map[string]interface{}{"error": err.Error()}
	}
	data, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return resp, result
}

// =============================================
//  AUTH TESTS
// =============================================

func TestRegisterFirstUserIsAdmin(t *testing.T) {
	base := startTestServer(t)

	resp, data := post(base+"/auth/register", map[string]interface{}{
		"email": "admin@crom.ai", "password": "pass123",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("register falhou: %d", resp.StatusCode)
	}
	if data["role"] != "admin" {
		t.Errorf("primeiro user deveria ser admin, é %v", data["role"])
	}
	if data["token"] == nil || data["token"] == "" {
		t.Error("token ausente")
	}
}

func TestRegisterSecondUserIsClient(t *testing.T) {
	base := startTestServer(t)

	post(base+"/auth/register", map[string]interface{}{
		"email": "admin@crom.ai", "password": "pass",
	})
	_, data := post(base+"/auth/register", map[string]interface{}{
		"email": "user@crom.ai", "password": "pass",
	})
	if data["role"] != "client" {
		t.Errorf("segundo user deveria ser client, é %v", data["role"])
	}
}

func TestRegisterDuplicate(t *testing.T) {
	base := startTestServer(t)

	post(base+"/auth/register", map[string]interface{}{
		"email": "dup@crom.ai", "password": "pass",
	})
	resp, _ := post(base+"/auth/register", map[string]interface{}{
		"email": "dup@crom.ai", "password": "pass",
	})
	if resp.StatusCode != 409 {
		t.Errorf("esperava 409, obteve %d", resp.StatusCode)
	}
}

func TestRegisterMissingFields(t *testing.T) {
	base := startTestServer(t)
	resp, _ := post(base+"/auth/register", map[string]interface{}{
		"email": "", "password": "",
	})
	if resp.StatusCode != 400 {
		t.Errorf("esperava 400, obteve %d", resp.StatusCode)
	}
}

func TestLoginSuccess(t *testing.T) {
	base := startTestServer(t)

	post(base+"/auth/register", map[string]interface{}{
		"email": "login@crom.ai", "password": "mypass",
	})
	resp, data := post(base+"/auth/login", map[string]interface{}{
		"email": "login@crom.ai", "password": "mypass",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("login falhou: %d", resp.StatusCode)
	}
	credits, ok := data["credits"].(float64)
	if !ok || credits != 10.0 {
		t.Errorf("créditos errados: %v", data["credits"])
	}
}

func TestLoginBadPassword(t *testing.T) {
	base := startTestServer(t)

	post(base+"/auth/register", map[string]interface{}{
		"email": "bad@crom.ai", "password": "correct",
	})
	resp, _ := post(base+"/auth/login", map[string]interface{}{
		"email": "bad@crom.ai", "password": "wrong",
	})
	if resp.StatusCode != 401 {
		t.Errorf("esperava 401, obteve %d", resp.StatusCode)
	}
}

func TestLoginNonExistent(t *testing.T) {
	base := startTestServer(t)
	resp, _ := post(base+"/auth/login", map[string]interface{}{
		"email": "ghost@crom.ai", "password": "pass",
	})
	if resp.StatusCode != 401 {
		t.Errorf("esperava 401, obteve %d", resp.StatusCode)
	}
}

func TestAuthMe(t *testing.T) {
	base := startTestServer(t)
	_, regData := post(base+"/auth/register", map[string]interface{}{
		"email": "me@crom.ai", "password": "pass",
	})
	token := regData["token"].(string)

	resp, data := get(base+"/auth/me", token)
	if resp.StatusCode != 200 {
		t.Fatalf("auth/me falhou: %d", resp.StatusCode)
	}
	if data["email"] != "me@crom.ai" {
		t.Errorf("email errado: %v", data["email"])
	}
}

// =============================================
//  BILLING TESTS
// =============================================

func TestBillingStatusAuth(t *testing.T) {
	base := startTestServer(t)

	resp, _ := get(base+"/billing/status", "")
	if resp.StatusCode != 401 {
		t.Errorf("esperava 401 sem token, obteve %d", resp.StatusCode)
	}

	_, regData := post(base+"/auth/register", map[string]interface{}{
		"email": "billing@crom.ai", "password": "pass",
	})
	token := regData["token"].(string)

	resp, data := get(base+"/billing/status", token)
	if resp.StatusCode != 200 {
		t.Fatalf("billing falhou: %d", resp.StatusCode)
	}
	if data["email"] != "billing@crom.ai" {
		t.Errorf("email errado: %v", data["email"])
	}
}

func TestBillingPurchase(t *testing.T) {
	base := startTestServer(t)

	_, regData := post(base+"/auth/register", map[string]interface{}{
		"email": "buyer@crom.ai", "password": "pass",
	})
	token := regData["token"].(string)

	resp, _ := authReq("POST", base+"/billing/purchase", token, map[string]interface{}{"amount": 50.0})
	if resp.StatusCode != 200 {
		t.Fatalf("purchase falhou: %d", resp.StatusCode)
	}

	_, data := get(base+"/billing/status", token)
	credits := data["credits"].(float64)
	if credits != 60.0 {
		t.Errorf("esperava 60.0, obteve %.2f", credits)
	}
}

// =============================================
//  MACHINES TESTS
// =============================================

func TestMachinesList(t *testing.T) {
	base := startTestServer(t)

	resp, err := http.Get(base + "/machines")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var machines []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&machines)

	if len(machines) < 3 {
		t.Errorf("esperava pelo menos 3 máquinas, obteve %d", len(machines))
	}
}

func TestMachineRent(t *testing.T) {
	base := startTestServer(t)

	_, regData := post(base+"/auth/register", map[string]interface{}{
		"email": "renter@crom.ai", "password": "pass",
	})
	token := regData["token"].(string)

	resp, _ := authReq("POST", base+"/machines/rent", token, map[string]interface{}{
		"machine_id": "crom-t4-01",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("rent falhou: %d", resp.StatusCode)
	}

	_, data := get(base+"/billing/status", token)
	credits := data["credits"].(float64)
	if credits != 9.75 {
		t.Errorf("esperava 9.75, obteve %.2f", credits)
	}
}

func TestMachineRentAlreadyRented(t *testing.T) {
	base := startTestServer(t)

	_, regData := post(base+"/auth/register", map[string]interface{}{
		"email": "renter2@crom.ai", "password": "pass",
	})
	token := regData["token"].(string)

	authReq("POST", base+"/machines/rent", token, map[string]interface{}{
		"machine_id": "crom-a100-01",
	})
	resp, _ := authReq("POST", base+"/machines/rent", token, map[string]interface{}{
		"machine_id": "crom-a100-01",
	})
	if resp.StatusCode != 409 {
		t.Errorf("esperava 409, obteve %d", resp.StatusCode)
	}
}

func TestMachineNotFound(t *testing.T) {
	base := startTestServer(t)

	_, regData := post(base+"/auth/register", map[string]interface{}{
		"email": "finder@crom.ai", "password": "pass",
	})
	token := regData["token"].(string)

	resp, _ := authReq("POST", base+"/machines/rent", token, map[string]interface{}{
		"machine_id": "nao-existe",
	})
	if resp.StatusCode != 404 {
		t.Errorf("esperava 404, obteve %d", resp.StatusCode)
	}
}

// =============================================
//  ADMIN TESTS
// =============================================

func TestAdminPlanCRUD(t *testing.T) {
	base := startTestServer(t)

	// Register admin (first user)
	_, regData := post(base+"/auth/register", map[string]interface{}{
		"email": "admin@crom.ai", "password": "admin",
	})
	token := regData["token"].(string)

	// Create plan
	resp, _ := authReq("POST", base+"/admin/plans", token, map[string]interface{}{
		"id": "start", "name": "Start", "vram": "6GB", "storage": "100GB",
		"price_hr": 0.30, "price_month": 29.90, "max_users": 50,
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create plan falhou: %d", resp.StatusCode)
	}

	// List plans
	resp, _ = get(base+"/admin/plans", token)
	if resp.StatusCode != 200 {
		t.Fatalf("list plans falhou: %d", resp.StatusCode)
	}

	// Get single plan
	resp, _ = get(base+"/admin/plans/start", token)
	if resp.StatusCode != 200 {
		t.Fatalf("get plan falhou: %d", resp.StatusCode)
	}

	// Delete plan
	resp, _ = authReq("DELETE", base+"/admin/plans/start", token, nil)
	if resp.StatusCode != 200 {
		t.Fatalf("delete plan falhou: %d", resp.StatusCode)
	}
}

func TestAdminPoolManagement(t *testing.T) {
	base := startTestServer(t)

	_, regData := post(base+"/auth/register", map[string]interface{}{
		"email": "admin@pool.ai", "password": "admin",
	})
	token := regData["token"].(string)

	// Create plan first
	authReq("POST", base+"/admin/plans", token, map[string]interface{}{
		"id": "pro", "name": "Pro", "vram": "24GB",
	})

	// Add pool entries
	authReq("POST", base+"/admin/pool/pro", token, map[string]interface{}{
		"priority": 1, "provider": "vastai", "label": "Vast T4", "address": "10.0.0.1:4422",
	})
	authReq("POST", base+"/admin/pool/pro", token, map[string]interface{}{
		"priority": 2, "provider": "runpod", "label": "RunPod", "address": "10.0.0.2:4422",
	})

	// List pool
	resp, _ := get(base+"/admin/pool/pro", token)
	if resp.StatusCode != 200 {
		t.Fatalf("list pool falhou: %d", resp.StatusCode)
	}
}

func TestAdminRejectsClient(t *testing.T) {
	base := startTestServer(t)

	// Admin
	post(base+"/auth/register", map[string]interface{}{
		"email": "admin@sec.ai", "password": "admin",
	})
	// Client (second user)
	_, clientData := post(base+"/auth/register", map[string]interface{}{
		"email": "client@sec.ai", "password": "client",
	})
	clientToken := clientData["token"].(string)

	// Client tries admin endpoint
	resp, _ := get(base+"/admin/plans", clientToken)
	if resp.StatusCode != 403 {
		t.Errorf("client acessou admin, esperava 403, obteve %d", resp.StatusCode)
	}
}

func TestAdminDashboard(t *testing.T) {
	base := startTestServer(t)

	_, regData := post(base+"/auth/register", map[string]interface{}{
		"email": "admin@dash.ai", "password": "admin",
	})
	token := regData["token"].(string)

	resp, data := get(base+"/admin/dashboard", token)
	if resp.StatusCode != 200 {
		t.Fatalf("dashboard falhou: %d", resp.StatusCode)
	}
	if data["users_total"] == nil {
		t.Error("dashboard sem users_total")
	}
}

// =============================================
//  CLIENT TESTS
// =============================================

func TestClientSubscribe(t *testing.T) {
	base := startTestServer(t)

	// Admin creates plan
	_, adminData := post(base+"/auth/register", map[string]interface{}{
		"email": "admin@sub.ai", "password": "admin",
	})
	adminToken := adminData["token"].(string)

	authReq("POST", base+"/admin/plans", adminToken, map[string]interface{}{
		"id": "basic", "name": "Basic", "price_hr": 0.10,
	})

	// Client subscribes
	_, clientData := post(base+"/auth/register", map[string]interface{}{
		"email": "client@sub.ai", "password": "client",
	})
	clientToken := clientData["token"].(string)

	resp, _ := authReq("POST", base+"/client/subscribe", clientToken, map[string]interface{}{
		"plan_id": "basic",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("subscribe falhou: %d", resp.StatusCode)
	}

	// Check subscription
	resp, data := get(base+"/client/subscription", clientToken)
	if resp.StatusCode != 200 {
		t.Fatalf("get subscription falhou: %d", resp.StatusCode)
	}
	plan := data["plan"].(map[string]interface{})
	if plan["name"] != "Basic" {
		t.Errorf("plano errado: %v", plan["name"])
	}
}

func TestClientPlansPublic(t *testing.T) {
	base := startTestServer(t)

	// Admin creates plan
	_, adminData := post(base+"/auth/register", map[string]interface{}{
		"email": "admin@pub.ai", "password": "admin",
	})
	adminToken := adminData["token"].(string)

	authReq("POST", base+"/admin/plans", adminToken, map[string]interface{}{
		"id": "vis", "name": "Visible", "vram": "8GB", "price_hr": 0.20,
	})

	// Client sees plans (no pool details)
	resp, _ := http.Get(base + "/client/plans")
	if resp.StatusCode != 200 {
		t.Fatalf("client plans falhou: %d", resp.StatusCode)
	}
}

var _ = os.TempDir // ensure os is used

