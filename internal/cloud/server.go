// Copyright (c) 2026 Crolab Contributors. All rights reserved.
// Licensed under the Crolab Sustainable License (CSL).
// Contact: mrj.crom@gmail.com
package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
	"github.com/crolab/core/web"
)

// --- Helpers ---

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, status int, msg string) {
	jsonResponse(w, status, map[string]string{"error": msg})
}

// sanitize strips dangerous chars from user input
func sanitize(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "\"", "")
	s = strings.ReplaceAll(s, ";", "")
	s = strings.ReplaceAll(s, "--", "")
	return s
}

func getAuthUser(r *http.Request) *DBUser {
	token := r.Header.Get("Authorization")
	if token == "" {
		return nil
	}
	u, err := DBGetUserByToken(token)
	if err != nil {
		return nil
	}
	return u
}

func requireAuth(w http.ResponseWriter, r *http.Request) *DBUser {
	user := getAuthUser(r)
	if user == nil {
		jsonError(w, 401, "token inválido")
		return nil
	}
	return user
}

func requireAdmin(w http.ResponseWriter, r *http.Request) *DBUser {
	user := requireAuth(w, r)
	if user == nil {
		return nil
	}
	if user.Role != "admin" {
		jsonError(w, 403, "acesso restrito a administradores")
		return nil
	}
	return user
}

// --- Rate Limiter ---

type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int           // max requests
	window   time.Duration // per window
}

type visitor struct {
	count   int
	resetAt time.Time
}

func newRateLimiter(rate int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}
	// Cleanup goroutine
	go func() {
		for {
			time.Sleep(window)
			rl.mu.Lock()
			now := time.Now()
			for ip, v := range rl.visitors {
				if now.After(v.resetAt) {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()
	return rl
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists || time.Now().After(v.resetAt) {
		rl.visitors[ip] = &visitor{count: 1, resetAt: time.Now().Add(rl.window)}
		return true
	}
	v.count++
	return v.count <= rl.rate
}

var limiter = newRateLimiter(60, time.Minute) // 60 req/min per IP

// --- Middleware Stack ---

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Security headers
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

// Rate limiting middleware
func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			ip = strings.Split(fwd, ",")[0]
		}
		if !limiter.allow(strings.TrimSpace(ip)) {
			w.Header().Set("Retry-After", "60")
			jsonError(w, 429, "rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Request logging middleware
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(sw, r)
		log.Printf("[%s] %s %s → %d (%s)",
			r.RemoteAddr, r.Method, r.URL.Path, sw.status, time.Since(start).Round(time.Microsecond))
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(s int) {
	w.status = s
	w.ResponseWriter.WriteHeader(s)
}

// =============================================
//  AUTH HANDLERS
// =============================================

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, 405, "POST only")
		return
	}
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	body.Email = sanitize(body.Email)

	if body.Email == "" || body.Password == "" {
		jsonError(w, 400, "email e password obrigatórios")
		return
	}

	// First user becomes admin
	role := "client"
	if DBUserCount() == 0 {
		role = "admin"
	}

	user, err := DBCreateUser(body.Email, body.Password, role)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			jsonError(w, 409, "email já registrado")
			return
		}
		jsonError(w, 500, err.Error())
		return
	}

	DBLogTransaction(user.ID, 10.0, "welcome", "Créditos de boas-vindas")
	log.Printf("☁️  Novo usuário: %s (role: %s)", body.Email, role)

	jsonResponse(w, 201, map[string]interface{}{
		"token":   user.Token,
		"role":    user.Role,
		"message": "Conta criada. 10 créditos de boas-vindas.",
	})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, 405, "POST only")
		return
	}
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	hash, err := DBGetPasswordHash(body.Email)
	if err != nil || !CheckPassword(hash, body.Password) {
		jsonError(w, 401, "credenciais inválidas")
		return
	}

	user, _ := DBGetUserByEmail(body.Email)
	jsonResponse(w, 200, map[string]interface{}{
		"token":   user.Token,
		"credits": user.Credits,
		"role":    user.Role,
	})
}

func handleLocalToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Permite origin localhost
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	// Puxa o token local do CLI
	viper.ReadInConfig()
	tok := viper.GetString("cloud_token")
	if tok == "" {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "not logged in locally"})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": tok})
}

func handleAuthMe(w http.ResponseWriter, r *http.Request) {
	user := requireAuth(w, r)
	if user == nil {
		return
	}
	jsonResponse(w, 200, map[string]interface{}{
		"email":   user.Email,
		"credits": user.Credits,
		"role":    user.Role,
	})
}

// =============================================
//  BILLING HANDLERS
// =============================================

func handleBillingStatus(w http.ResponseWriter, r *http.Request) {
	user := requireAuth(w, r)
	if user == nil {
		return
	}
	jsonResponse(w, 200, map[string]interface{}{
		"email":   user.Email,
		"credits": user.Credits,
	})
}

func handleBillingPurchase(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, 405, "POST only")
		return
	}
	user := requireAuth(w, r)
	if user == nil {
		return
	}

	var body struct {
		Amount float64 `json:"amount"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	if body.Amount <= 0 {
		jsonError(w, 400, "amount deve ser positivo")
		return
	}

	credits, _ := DBUpdateCredits(user.ID, body.Amount)
	DBLogTransaction(user.ID, body.Amount, "purchase", fmt.Sprintf("Compra de %.2f créditos", body.Amount))

	jsonResponse(w, 200, map[string]interface{}{
		"credits": credits,
		"message": fmt.Sprintf("%.2f créditos adicionados", body.Amount),
	})
}

func handleBillingTransactions(w http.ResponseWriter, r *http.Request) {
	user := requireAuth(w, r)
	if user == nil {
		return
	}
	txs, _ := DBListTransactions(user.ID)
	jsonResponse(w, 200, txs)
}

// =============================================
//  MACHINES HANDLERS (Public)
// =============================================

func handleMachinesList(w http.ResponseWriter, r *http.Request) {
	machines, _ := DBListMachines()
	if machines == nil {
		machines = []DBMachine{}
	}
	jsonResponse(w, 200, machines)
}

func handleMachineRent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, 405, "POST only")
		return
	}
	user := requireAuth(w, r)
	if user == nil {
		return
	}

	var body struct {
		MachineID string `json:"machine_id"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	machines, _ := DBListMachines()
	var machine *DBMachine
	for i := range machines {
		if machines[i].ID == body.MachineID {
			machine = &machines[i]
			break
		}
	}

	if machine == nil {
		jsonError(w, 404, "máquina não encontrada")
		return
	}
	if machine.Status != "available" {
		jsonError(w, 409, "máquina já alugada")
		return
	}
	if user.Credits < machine.PriceHr {
		jsonError(w, 402, fmt.Sprintf("créditos insuficientes (precisa: %.2f, tem: %.2f)", machine.PriceHr, user.Credits))
		return
	}

	DBUpdateCredits(user.ID, -machine.PriceHr)
	DBLogTransaction(user.ID, -machine.PriceHr, "rent", fmt.Sprintf("Aluguel %s", machine.Name))
	db.Exec("UPDATE machines SET status='rented', rented_by=? WHERE id=?", user.Email, machine.ID)

	log.Printf("☁️  Máquina %s alugada por %s", machine.Name, user.Email)

	updatedUser, _ := DBGetUserByToken(user.Token)
	jsonResponse(w, 200, map[string]interface{}{
		"machine": machine,
		"message": fmt.Sprintf("Máquina %s ativada. Use: crolab config add %s %s", machine.Name, machine.Name, machine.Address),
		"credits": updatedUser.Credits,
	})
}

// =============================================
//  ADMIN HANDLERS
// =============================================

func handleAdminPlans(w http.ResponseWriter, r *http.Request) {
	if requireAdmin(w, r) == nil {
		return
	}

	switch r.Method {
	case http.MethodGet:
		plans, _ := DBListPlans()
		if plans == nil {
			plans = []DBPlan{}
		}
		jsonResponse(w, 200, plans)

	case http.MethodPost:
		var p DBPlan
		json.NewDecoder(r.Body).Decode(&p)
		if p.ID == "" || p.Name == "" {
			jsonError(w, 400, "id e name obrigatórios")
			return
		}
		if err := DBCreatePlan(p); err != nil {
			jsonError(w, 409, err.Error())
			return
		}
		log.Printf("📋 Plano criado: %s (%s)", p.Name, p.ID)
		jsonResponse(w, 201, p)

	default:
		jsonError(w, 405, "GET ou POST")
	}
}

func handleAdminPlanByID(w http.ResponseWriter, r *http.Request) {
	if requireAdmin(w, r) == nil {
		return
	}

	// Extract plan ID from path: /admin/plans/{id}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		jsonError(w, 400, "plan ID obrigatório")
		return
	}
	planID := parts[3]

	switch r.Method {
	case http.MethodGet:
		p, err := DBGetPlan(planID)
		if err != nil {
			jsonError(w, 404, "plano não encontrado")
			return
		}
		jsonResponse(w, 200, p)

	case http.MethodPut:
		var p DBPlan
		json.NewDecoder(r.Body).Decode(&p)
		p.ID = planID
		DBUpdatePlan(p)
		jsonResponse(w, 200, map[string]string{"status": "updated"})

	case http.MethodDelete:
		DBDeletePlan(planID)
		jsonResponse(w, 200, map[string]string{"status": "deleted"})

	default:
		jsonError(w, 405, "GET, PUT ou DELETE")
	}
}

func pushWebhookMsg(title, desc string) {
	webhookUrl := os.Getenv("CROLAB_WEBHOOK_URL")
	if webhookUrl == "" {
		return
	}
	payload, _ := json.Marshal(map[string]string{"title": title, "text": desc})
	http.Post(webhookUrl, "application/json", bytes.NewBuffer(payload))
}

func handleAdminSpread(w http.ResponseWriter, r *http.Request) {
	if requireAdmin(w, r) == nil {
		return
	}
	machines, _ := DBListMachines()
	var totalCost, totalBilled float64

	for _, m := range machines {
		totalCost += m.ProviderCostHr
		totalBilled += m.PriceHr
	}

	jsonResponse(w, 200, map[string]interface{}{
		"provider_cost": totalCost,
		"client_billed": totalBilled,
		"spread_profit": totalBilled - totalCost,
		"gross_margin":  ((totalBilled - totalCost) / totalBilled) * 100,
	})
}

func handleAdminPool(w http.ResponseWriter, r *http.Request) {
	if requireAdmin(w, r) == nil {
		return
	}

	// /admin/pool/{planID}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		jsonError(w, 400, "plan ID obrigatório")
		return
	}
	planID := parts[3]

	switch r.Method {
	case http.MethodGet:
		entries, _ := DBListPool(planID)
		if entries == nil {
			entries = []DBPoolEntry{}
		}
		jsonResponse(w, 200, entries)

	case http.MethodPost:
		var e DBPoolEntry
		json.NewDecoder(r.Body).Decode(&e)
		e.PlanID = planID
		if e.Address == "" {
			jsonError(w, 400, "address obrigatório")
			return
		}
		DBAddPoolEntry(e)
		log.Printf("📋 Pool entry adicionada ao plano %s: priority=%d %s", planID, e.Priority, e.Provider)
		jsonResponse(w, 201, e)

	case http.MethodDelete:
		// /admin/pool/{planID}/{entryID}
		if len(parts) < 5 {
			jsonError(w, 400, "entry ID obrigatório")
			return
		}
		entryID, _ := strconv.Atoi(parts[4])
		DBDeletePoolEntry(entryID)
		jsonResponse(w, 200, map[string]string{"status": "deleted"})

	default:
		jsonError(w, 405, "GET, POST ou DELETE")
	}
}

func handleAdminMachines(w http.ResponseWriter, r *http.Request) {
	if requireAdmin(w, r) == nil {
		return
	}

	switch r.Method {
	case http.MethodGet:
		machines, _ := DBListMachines()
		if machines == nil {
			machines = []DBMachine{}
		}
		jsonResponse(w, 200, machines)

	case http.MethodPost:
		var m DBMachine
		json.NewDecoder(r.Body).Decode(&m)
		if m.ID == "" || m.Name == "" {
			jsonError(w, 400, "id e name obrigatórios")
			return
		}
		DBCreateMachine(m)
		log.Printf("🖥️  Máquina adicionada: %s (%s)", m.Name, m.GPU)
		jsonResponse(w, 201, m)

	case http.MethodDelete:
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 4 {
			jsonError(w, 400, "machine ID obrigatório")
			return
		}
		DBDeleteMachine(parts[3])
		jsonResponse(w, 200, map[string]string{"status": "deleted"})

	default:
		jsonError(w, 405, "GET, POST ou DELETE")
	}
}

func handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	if requireAdmin(w, r) == nil {
		return
	}

	switch r.Method {
	case http.MethodGet:
		users, _ := DBListUsers()
		jsonResponse(w, 200, users)

	case http.MethodPut:
		// /admin/users/{id}/credits or /admin/users/{id}/role
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 5 {
			jsonError(w, 400, "user ID e ação obrigatórios")
			return
		}
		userID, _ := strconv.Atoi(parts[3])
		action := parts[4]

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		switch action {
		case "credits":
			credits, _ := body["credits"].(float64)
			DBUpdateUserCredits(userID, credits)
			jsonResponse(w, 200, map[string]string{"status": "credits updated"})
		case "role":
			role, _ := body["role"].(string)
			DBUpdateUserRole(userID, role)
			jsonResponse(w, 200, map[string]string{"status": "role updated"})
		default:
			jsonError(w, 400, "ação inválida: use 'credits' ou 'role'")
		}

	default:
		jsonError(w, 405, "GET ou PUT")
	}
}

func handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	if requireAdmin(w, r) == nil {
		return
	}

	users, _ := DBListUsers()
	plans, _ := DBListPlans()
	machines, _ := DBListMachines()

	online := 0
	for _, m := range machines {
		if m.Status == "available" || m.Status == "rented" {
			online++
		}
	}

	jsonResponse(w, 200, map[string]interface{}{
		"users_total":    len(users),
		"plans_total":    len(plans),
		"machines_total": len(machines),
		"machines_online": online,
	})
}

// =============================================
//  OBSERVABILITY / PROMETHEUS
// =============================================

func handlePrometheusMetrics(w http.ResponseWriter, r *http.Request) {
	users, _ := DBListUsers()
	plans, _ := DBListPlans()
	machines, _ := DBListMachines()
	
	online := 0
	for _, m := range machines {
		if m.Status == "available" || m.Status == "rented" {
			online++
		}
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.WriteHeader(200)

	fmt.Fprintf(w, "# HELP crolab_users_total Total de usuários registrados na plataforma.\n")
	fmt.Fprintf(w, "# TYPE crolab_users_total gauge\n")
	fmt.Fprintf(w, "crolab_users_total %d\n\n", len(users))

	fmt.Fprintf(w, "# HELP crolab_plans_total Total de planos disponíveis.\n")
	fmt.Fprintf(w, "# TYPE crolab_plans_total gauge\n")
	fmt.Fprintf(w, "crolab_plans_total %d\n\n", len(plans))

	fmt.Fprintf(w, "# HELP crolab_machines_total Total de máquinas na rede.\n")
	fmt.Fprintf(w, "# TYPE crolab_machines_total gauge\n")
	fmt.Fprintf(w, "crolab_machines_total %d\n\n", len(machines))

	fmt.Fprintf(w, "# HELP crolab_machines_online_total Total de máquinas disponíveis ou rodando jobs.\n")
	fmt.Fprintf(w, "# TYPE crolab_machines_online_total gauge\n")
	fmt.Fprintf(w, "crolab_machines_online_total %d\n\n", online)
}

// =============================================
//  CLIENT HANDLERS
// =============================================

func handleClientPlans(w http.ResponseWriter, r *http.Request) {
	plans, _ := DBListPlans()
	// Filter: only active, no pool details
	var filtered []map[string]interface{}
	for _, p := range plans {
		if p.Active {
			filtered = append(filtered, map[string]interface{}{
				"id": p.ID, "name": p.Name, "vram": p.VRAM,
				"storage": p.Storage, "price_hr": p.PriceHr,
				"price_month": p.PriceMonth,
			})
		}
	}
	if filtered == nil {
		filtered = []map[string]interface{}{}
	}
	jsonResponse(w, 200, filtered)
}

func handleClientSubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, 405, "POST only")
		return
	}
	user := requireAuth(w, r)
	if user == nil {
		return
	}

	var body struct {
		PlanID string `json:"plan_id"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	plan, err := DBGetPlan(body.PlanID)
	if err != nil {
		jsonError(w, 404, "plano não encontrado")
		return
	}

	if user.Credits < plan.PriceHr {
		jsonError(w, 402, "créditos insuficientes")
		return
	}

	DBSubscribe(user.ID, body.PlanID)
	DBLogTransaction(user.ID, 0, "subscribe", fmt.Sprintf("Assinatura do plano %s", plan.Name))

	jsonResponse(w, 200, map[string]interface{}{
		"plan":    plan.Name,
		"message": "Plano ativado com sucesso",
	})
}

func handleClientSubscription(w http.ResponseWriter, r *http.Request) {
	user := requireAuth(w, r)
	if user == nil {
		return
	}

	if r.Method == http.MethodDelete {
		DBUnsubscribe(user.ID)
		jsonResponse(w, 200, map[string]string{"status": "unsubscribed"})
		return
	}

	plan, err := DBGetSubscription(user.ID)
	if err != nil {
		jsonResponse(w, 200, map[string]interface{}{"plan": nil, "message": "sem plano ativo"})
		return
	}
	jsonResponse(w, 200, map[string]interface{}{"plan": plan})
}

// =============================================
//  CLIENT MACHINES
// =============================================

func handleClientMachines(w http.ResponseWriter, r *http.Request) {
	user := requireAuth(w, r)
	if user == nil {
		return
	}

	switch r.Method {
	case http.MethodGet:
		machines, _ := DBListUserMachines(user.ID)
		if machines == nil {
			machines = []DBUserMachine{}
		}
		jsonResponse(w, 200, machines)

	case http.MethodPost:
		var body struct {
			Name     string `json:"name"`
			Address  string `json:"address"`
			Token    string `json:"token"`
			Provider string `json:"provider"`
			Priority int    `json:"priority"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		if body.Address == "" || body.Name == "" {
			jsonError(w, 400, "name e address obrigatórios")
			return
		}
		if body.Provider == "" {
			body.Provider = "personal"
		}
		if body.Priority == 0 {
			body.Priority = 1
		}
		DBAddUserMachine(user.ID, body.Name, body.Address, body.Token, body.Provider, body.Priority)
		log.Printf("🔗 Máquina pessoal conectada: %s → %s (user: %s)", body.Name, body.Address, user.Email)
		jsonResponse(w, 201, map[string]string{"status": "connected"})

	case http.MethodDelete:
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 4 {
			jsonError(w, 400, "machine ID obrigatório")
			return
		}
		machineID, _ := strconv.Atoi(parts[3])
		DBDeleteUserMachine(user.ID, machineID)
		jsonResponse(w, 200, map[string]string{"status": "disconnected"})

	default:
		jsonError(w, 405, "GET, POST ou DELETE")
	}
}

// =============================================
//  CLIENT JOBS
// =============================================

func handleClientJobs(w http.ResponseWriter, r *http.Request) {
	user := requireAuth(w, r)
	if user == nil {
		return
	}

	jobs, _ := DBListJobs(user.ID)
	if jobs == nil {
		jobs = []DBJob{}
	}
	jsonResponse(w, 200, jobs)
}

func handleClientRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, 405, "POST only")
		return
	}
	user := requireAuth(w, r)
	if user == nil {
		return
	}

	var body struct {
		PlanID    string `json:"plan_id"`
		MachineID string `json:"machine_id"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	var nodes []map[string]string

	if body.MachineID != "" {
		// Busca máquina pessoal ou pública para rodar
		myMachines, _ := DBListUserMachines(user.ID)
		for _, m := range myMachines {
			if m.Name == body.MachineID {
				nodes = append(nodes, map[string]string{
					"address": m.Address,
					"token":   m.Token,
					"machine": "personal:" + m.Name,
				})
				break
			}
		}
		if len(nodes) == 0 {
			allMachines, _ := DBListMachines()
			for _, m := range allMachines {
				if m.ID == body.MachineID {
					nodes = append(nodes, map[string]string{
						"address": m.Address,
						"token":   "",
						"machine": m.ID,
					})
					break
				}
			}
		}
		if len(nodes) == 0 {
			jsonError(w, 404, "Máquina indisponível")
			return
		}

	} else if body.PlanID != "" {
		pool, err := DBListPool(body.PlanID)
		if err != nil || len(pool) == 0 {
			jsonError(w, 503, "Sem GPUs disponíveis neste pool no momento")
			return
		}
		
		for _, p := range pool {
			nodes = append(nodes, map[string]string{
				"address": p.Address,
				"token":   p.Token,
				"machine": p.Provider + ":" + p.Label,
			})
		}

		plan, _ := DBGetPlan(body.PlanID)
		if plan != nil && plan.PriceHr > 0 {
			if user.Credits < plan.PriceHr {
				jsonError(w, 402, "Créditos insuficientes")
				return
			}
			DBUpdateUserCredits(user.ID, user.Credits - plan.PriceHr)
			DBLogTransaction(user.ID, -plan.PriceHr, "job", "Cobrança antecipada "+body.PlanID)
		}

	} else {
		jsonError(w, 400, "plan_id ou machine_id obrigatório")
		return
	}

	jobID := GenerateToken()[:16]
	job := DBJob{
		ID:          jobID,
		UserID:      user.ID,
		PlanID:      body.PlanID,
		MachineUsed: nodes[0]["machine"],
		Status:      "running",
	}

	DBCreateJob(job)

	jsonResponse(w, 201, map[string]interface{}{
		"job_id":  jobID,
		"nodes":   nodes,
		"status":  "running",
		"message": "Enviando job diretamente à GPU associada...",
	})
}

// =============================================
//  ADMIN EXTENDED
// =============================================

func handleAdminUserByID(w http.ResponseWriter, r *http.Request) {
	if requireAdmin(w, r) == nil {
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		jsonError(w, 400, "user ID obrigatório")
		return
	}
	userID, _ := strconv.Atoi(parts[3])

	// Check for sub-resource: /admin/users/{id}/credits or /admin/users/{id}/role
	if len(parts) >= 5 {
		action := parts[4]
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		switch action {
		case "credits":
			credits, _ := body["credits"].(float64)
			DBUpdateUserCredits(userID, credits)
			jsonResponse(w, 200, map[string]string{"status": "credits updated"})
		case "role":
			role, _ := body["role"].(string)
			DBUpdateUserRole(userID, role)
			jsonResponse(w, 200, map[string]string{"status": "role updated"})
		default:
			jsonError(w, 400, "ação inválida")
		}
		return
	}

	switch r.Method {
	case http.MethodGet:
		u, err := DBGetUserByID(userID)
		if err != nil {
			jsonError(w, 404, "usuário não encontrado")
			return
		}
		jsonResponse(w, 200, u)

	case http.MethodDelete:
		DBDeleteUser(userID)
		jsonResponse(w, 200, map[string]string{"status": "disabled"})

	default:
		jsonError(w, 405, "GET ou DELETE")
	}
}

func handleAdminLogs(w http.ResponseWriter, r *http.Request) {
	if requireAdmin(w, r) == nil {
		return
	}

	// Return last 100 transactions as audit log
	rows, err := db.Query(
		`SELECT t.id, t.user_id, u.email, t.amount, t.type, t.description, t.created_at
		 FROM transactions t JOIN users u ON t.user_id = u.id
		 ORDER BY t.id DESC LIMIT 100`,
	)
	if err != nil {
		jsonResponse(w, 200, []interface{}{})
		return
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var id, userID int
		var email, txType, desc, createdAt string
		var amount float64
		rows.Scan(&id, &userID, &email, &amount, &txType, &desc, &createdAt)
		logs = append(logs, map[string]interface{}{
			"id": id, "user_id": userID, "email": email,
			"amount": amount, "type": txType, "description": desc, "created_at": createdAt,
		})
	}
	if logs == nil {
		logs = []map[string]interface{}{}
	}
	jsonResponse(w, 200, logs)
}

func handleAdminMachineByID(w http.ResponseWriter, r *http.Request) {
	if requireAdmin(w, r) == nil {
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		jsonError(w, 400, "machine ID obrigatório")
		return
	}
	machineID := parts[3]

	switch r.Method {
	case http.MethodPut:
		var m DBMachine
		json.NewDecoder(r.Body).Decode(&m)
		m.ID = machineID
		DBUpdateMachine(m)
		jsonResponse(w, 200, map[string]string{"status": "updated"})

	case http.MethodDelete:
		DBDeleteMachine(machineID)
		jsonResponse(w, 200, map[string]string{"status": "deleted"})

	default:
		jsonError(w, 405, "PUT ou DELETE")
	}
}

// =============================================
//  SERVER STARTUP
// =============================================

// SeedDefaultMachines populates the machines table with default entries if empty.
func SeedDefaultMachines() {
	machines, _ := DBListMachines()
	if len(machines) > 0 {
		return // Already seeded
	}

	defaults := []DBMachine{
		{ID: "crom-a100-01", Name: "A100-Brazil-01", GPU: "A100", VRAM: "80GB", PriceHr: 1.50, Status: "available", Provider: "crom"},
		{ID: "crom-a100-02", Name: "A100-Brazil-02", GPU: "A100", VRAM: "80GB", PriceHr: 1.50, Status: "available", Provider: "crom"},
		{ID: "crom-4090-01", Name: "RTX4090-EU-01", GPU: "RTX 4090", VRAM: "24GB", PriceHr: 0.60, Status: "available", Provider: "crom"},
		{ID: "crom-t4-01", Name: "T4-US-01", GPU: "T4", VRAM: "16GB", PriceHr: 0.25, Status: "available", Provider: "crom"},
	}
	for _, m := range defaults {
		DBCreateMachine(m)
	}
	log.Println("🖥️  Máquinas padrão criadas (4)")
}

// BuildMux creates the HTTP mux with all API routes + optional static files.
func BuildMux(webDir string) http.Handler {
	mux := http.NewServeMux()

	// Auth
	mux.HandleFunc("/auth/register", handleRegister)
	mux.HandleFunc("/auth/login", handleLogin)
	mux.HandleFunc("/auth/me", handleAuthMe)
	mux.HandleFunc("/auth/local-token", handleLocalToken)

	// Billing
	mux.HandleFunc("/billing/status", handleBillingStatus)
	mux.HandleFunc("/billing/purchase", handleBillingPurchase)
	mux.HandleFunc("/billing/transactions", handleBillingTransactions)

	// Machines (public)
	mux.HandleFunc("/machines", handleMachinesList)
	mux.HandleFunc("/machines/rent", handleMachineRent)

	// Admin endpoints
	mux.HandleFunc("/admin/plans", handleAdminPlans)
	mux.HandleFunc("/admin/plans/", handleAdminPlanByID)
	mux.HandleFunc("/admin/pool/", handleAdminPool)
	mux.HandleFunc("/admin/providers/sync", handleAdminProvidersSync)
	mux.HandleFunc("/admin/machines", handleAdminMachines)
	mux.HandleFunc("/admin/machines/", handleAdminMachineByID)
	mux.HandleFunc("/admin/users", handleAdminUsers)
	mux.HandleFunc("/admin/users/", handleAdminUserByID)
	mux.HandleFunc("/admin/dashboard", handleAdminDashboard)
	mux.HandleFunc("/admin/spread", handleAdminSpread)
	mux.HandleFunc("/admin/logs", handleAdminLogs)

	// Client endpoints
	mux.HandleFunc("/client/plans", handleClientPlans)
	mux.HandleFunc("/client/subscribe", handleClientSubscribe)
	mux.HandleFunc("/client/subscription", handleClientSubscription)
	mux.HandleFunc("/client/machines", handleClientMachines)
	mux.HandleFunc("/client/machines/", handleClientMachines)
	mux.HandleFunc("/client/jobs", handleClientJobs)
	mux.HandleFunc("/client/run", handleClientRun)
	
	// Tensors P2P (Chunked Streaming API)
	mux.HandleFunc("/tensors/upload", handleTensorUpload)
	mux.HandleFunc("/tensors/download/", handleTensorDownload)

	// Observabilidade
	mux.HandleFunc("/metrics", handlePrometheusMetrics)

	// Serve frontend
	if webDir != "" {
		fs := http.FileServer(http.Dir(webDir))
		mux.Handle("/", fs)
		
		// Expose SDK global
		sdkPath := filepath.Join(filepath.Dir(webDir), "sdk")
		if stat, err := os.Stat(sdkPath); err == nil && stat.IsDir() {
			mux.Handle("/sdk/", http.StripPrefix("/sdk/", http.FileServer(http.Dir(sdkPath))))
		}
	} else {
		// Use embedded FS globally
		embedFS, err := fs.Sub(web.StaticFiles, ".")
		if err == nil {
			mux.Handle("/", http.FileServer(http.FS(embedFS)))
		}
	}

	// Middleware stack: logging → rate limit → security → CORS → routes
	return loggingMiddleware(rateLimitMiddleware(securityHeaders(corsMiddleware(mux))))
}

// StartCloudServer starts the Crom Cloud REST API with SQLite persistence.
func StartCloudServer(port string, webDir string, tlsCert string, tlsKey string) error {
	// Initialize DB
	if db == nil {
		if err := InitDB("crolab.db"); err != nil {
			return fmt.Errorf("falha ao iniciar DB: %w", err)
		}
		SeedDefaultMachines()
	}

	handler := BuildMux(webDir)

	protocol := "http"
	if tlsCert != "" && tlsKey != "" {
		protocol = "https"
	}

	if webDir != "" {
		log.Printf("🌐 Frontend em %s://localhost%s", protocol, port)
	}
	log.Printf("☁️  Crom Cloud API em %s://localhost%s", protocol, port)

	if protocol == "https" {
		return http.ListenAndServeTLS(port, tlsCert, tlsKey, handler)
	}
	return http.ListenAndServe(port, handler)
}

func handleAdminProvidersSync(w http.ResponseWriter, r *http.Request) {
	if requireAdmin(w, r) == nil {
		return
	}

	if r.Method != http.MethodPost {
		jsonError(w, 405, "Requer POST")
		return
	}

	total, err := SyncVastAIOffers()
	if err != nil {
		jsonError(w, 500, err.Error())
		return
	}

	jsonResponse(w, 200, map[string]interface{}{
		"status": "success",
		"message": fmt.Sprintf("Sincronização Vast.AI Mestre! %d máquinas sugadas para sua Cloud Local.", total),
		"count": total,
	})
}

// --- Tensor P2P Handlers ---
func handleTensorUpload(w http.ResponseWriter, r *http.Request) {
	if requireAuth(w, r) == nil {
		return
	}
	if r.Method != http.MethodPost {
		jsonError(w, http.StatusMethodNotAllowed, "Requer POST")
		return
	}

	err := r.ParseMultipartForm(500 << 20) // Limit to 500MB in-memory per chunk segment
	if err != nil {
		jsonError(w, http.StatusBadRequest, "Arquivo mutio grande ou malformado")
		return
	}

	file, header, err := r.FormFile("tensor")
	if err != nil {
		jsonError(w, http.StatusBadRequest, "Falha ao ler tensor")
		return
	}
	defer file.Close()

	os.MkdirAll(".crolab/tensors", 0755)
	destPath := filepath.Join(".crolab/tensors", sanitize(header.Filename))

	out, err := os.Create(destPath)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Erro de I/O no nó")
		return
	}
	defer out.Close()

	// chunked buffering auto
	_, err = out.ReadFrom(file)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Erro durante file-streaming")
		return
	}

	jsonResponse(w, 201, map[string]string{
		"status": "success",
		"message": "Tensor '" + header.Filename + "' sincronizado com o Nó P2P.",
	})
}

func handleTensorDownload(w http.ResponseWriter, r *http.Request) {
	if requireAuth(w, r) == nil {
		return
	}
	if r.Method != http.MethodGet {
		jsonError(w, http.StatusMethodNotAllowed, "Requer GET")
		return
	}
	filename := strings.TrimPrefix(r.URL.Path, "/tensors/download/")
	filename = sanitize(filename)
	if filename == "" {
		jsonError(w, http.StatusBadRequest, "Hash ou nome de Tensor inválido")
		return
	}

	targetPath := filepath.Join(".crolab/tensors", filename)
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		jsonError(w, http.StatusNotFound, "Tensor não encontrado neste Nó")
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	http.ServeFile(w, r, targetPath)
}
