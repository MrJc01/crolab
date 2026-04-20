// Copyright (c) 2026 Crolab Contributors. All rights reserved.
// Licensed under the Crolab Sustainable License (CSL).
// Contact: mrj.crom@gmail.com
package cloud

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

var db *sql.DB

// InitDB initializes SQLite database and creates tables.
func InitDB(path string) error {
	var err error
	db, err = sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("db open: %w", err)
	}

	// WAL mode for better concurrency
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA foreign_keys=ON")

	return migrate()
}

func migrate() error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS kv_settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)`,

		`CREATE TABLE IF NOT EXISTS users (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			email      TEXT    UNIQUE NOT NULL,
			password   TEXT    NOT NULL,
			token      TEXT    UNIQUE NOT NULL,
			credits    REAL    DEFAULT 10.0,
			role       TEXT    DEFAULT 'client',
			created_at TEXT    DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS plans (
			id          TEXT PRIMARY KEY,
			name        TEXT NOT NULL,
			vram        TEXT DEFAULT '',
			storage     TEXT DEFAULT '',
			price_hr    REAL DEFAULT 0.0,
			price_month REAL DEFAULT 0.0,
			max_users   INTEGER DEFAULT 100,
			active      INTEGER DEFAULT 1,
			created_at  TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS pool_entries (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			plan_id    TEXT    NOT NULL,
			priority   INTEGER NOT NULL,
			provider   TEXT    DEFAULT '',
			label      TEXT    DEFAULT '',
			machine_id TEXT    DEFAULT '',
			address    TEXT    NOT NULL,
			token      TEXT    DEFAULT '',
			FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS machines (
			id        TEXT PRIMARY KEY,
			name      TEXT NOT NULL,
			gpu       TEXT DEFAULT '',
			vram      TEXT DEFAULT '',
			price_hr  REAL DEFAULT 0.0,
			status    TEXT DEFAULT 'available',
			address   TEXT DEFAULT '',
			provider  TEXT DEFAULT '',
			provider_cost_hr REAL DEFAULT 0.0,
			rented_by TEXT DEFAULT '',
			created_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS subscriptions (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id    INTEGER NOT NULL,
			plan_id    TEXT    NOT NULL,
			active     INTEGER DEFAULT 1,
			started_at TEXT    DEFAULT (datetime('now')),
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (plan_id) REFERENCES plans(id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_machines (
			id       INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id  INTEGER NOT NULL,
			name     TEXT    NOT NULL,
			address  TEXT    NOT NULL,
			token    TEXT    DEFAULT '',
			provider TEXT    DEFAULT 'personal',
			priority INTEGER DEFAULT 1,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS jobs (
			id           TEXT PRIMARY KEY,
			user_id      INTEGER NOT NULL,
			plan_id      TEXT    DEFAULT '',
			machine_used TEXT    DEFAULT '',
			status       TEXT    DEFAULT 'queued',
			duration_s   REAL    DEFAULT 0,
			cost         REAL    DEFAULT 0,
			created_at   TEXT    DEFAULT (datetime('now')),
			finished_at  TEXT    DEFAULT '',
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id     INTEGER NOT NULL,
			amount      REAL    NOT NULL,
			type        TEXT    NOT NULL,
			description TEXT    DEFAULT '',
			created_at  TEXT    DEFAULT (datetime('now')),
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
	}

	for _, t := range tables {
		if _, err := db.Exec(t); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}

	log.Println("🗄️  SQLite migrado com sucesso")
	return nil
}

// --- User Operations ---

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func GenerateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

type DBUser struct {
	ID        int
	Email     string
	Token     string
	Credits   float64
	Role      string
	CreatedAt string
}

func DBCreateUser(email, password, role string) (*DBUser, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	token := GenerateToken()

	_, err = db.Exec(
		"INSERT INTO users (email, password, token, role) VALUES (?, ?, ?, ?)",
		email, hash, token, role,
	)
	if err != nil {
		return nil, err
	}

	return DBGetUserByEmail(email)
}

func DBGetUserByEmail(email string) (*DBUser, error) {
	u := &DBUser{}
	err := db.QueryRow(
		"SELECT id, email, token, credits, role, created_at FROM users WHERE email = ?",
		email,
	).Scan(&u.ID, &u.Email, &u.Token, &u.Credits, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func DBGetUserByToken(token string) (*DBUser, error) {
	u := &DBUser{}
	err := db.QueryRow(
		"SELECT id, email, token, credits, role, created_at FROM users WHERE token = ?",
		token,
	).Scan(&u.ID, &u.Email, &u.Token, &u.Credits, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func DBGetPasswordHash(email string) (string, error) {
	var hash string
	err := db.QueryRow("SELECT password FROM users WHERE email = ?", email).Scan(&hash)
	return hash, err
}

func DBUpdateCredits(userID int, delta float64) (float64, error) {
	_, err := db.Exec("UPDATE users SET credits = credits + ? WHERE id = ?", delta, userID)
	if err != nil {
		return 0, err
	}
	var credits float64
	db.QueryRow("SELECT credits FROM users WHERE id = ?", userID).Scan(&credits)
	return credits, nil
}

func DBUserCount() int {
	var c int
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&c)
	return c
}

func DBListUsers() ([]DBUser, error) {
	rows, err := db.Query("SELECT id, email, token, credits, role, created_at FROM users ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []DBUser
	for rows.Next() {
		var u DBUser
		rows.Scan(&u.ID, &u.Email, &u.Token, &u.Credits, &u.Role, &u.CreatedAt)
		users = append(users, u)
	}
	return users, nil
}

func DBUpdateUserCredits(userID int, newCredits float64) error {
	_, err := db.Exec("UPDATE users SET credits = ? WHERE id = ?", newCredits, userID)
	return err
}

func DBUpdateUserRole(userID int, role string) error {
	_, err := db.Exec("UPDATE users SET role = ? WHERE id = ?", role, userID)
	return err
}

// --- Plan Operations ---

type DBPlan struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	VRAM       string      `json:"vram"`
	Storage    string      `json:"storage"`
	PriceHr    float64     `json:"price_hr"`
	PriceMonth float64     `json:"price_month"`
	MaxUsers   int         `json:"max_users"`
	Active     bool        `json:"active"`
	Pool       []DBPoolEntry `json:"pool,omitempty"`
}

type DBPoolEntry struct {
	ID        int    `json:"id"`
	PlanID    string `json:"plan_id"`
	Priority  int    `json:"priority"`
	Provider  string `json:"provider"`
	Label     string `json:"label"`
	MachineID string `json:"machine_id"`
	Address   string `json:"address"`
	Token     string `json:"token,omitempty"`
}

func DBCreatePlan(p DBPlan) error {
	_, err := db.Exec(
		"INSERT INTO plans (id, name, vram, storage, price_hr, price_month, max_users) VALUES (?, ?, ?, ?, ?, ?, ?)",
		p.ID, p.Name, p.VRAM, p.Storage, p.PriceHr, p.PriceMonth, p.MaxUsers,
	)
	return err
}

func DBListPlans() ([]DBPlan, error) {
	rows, err := db.Query("SELECT id, name, vram, storage, price_hr, price_month, max_users, active FROM plans ORDER BY price_hr")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []DBPlan
	for rows.Next() {
		var p DBPlan
		var active int
		rows.Scan(&p.ID, &p.Name, &p.VRAM, &p.Storage, &p.PriceHr, &p.PriceMonth, &p.MaxUsers, &active)
		p.Active = active == 1
		plans = append(plans, p)
	}
	return plans, nil
}

func DBGetPlan(id string) (*DBPlan, error) {
	p := &DBPlan{}
	var active int
	err := db.QueryRow(
		"SELECT id, name, vram, storage, price_hr, price_month, max_users, active FROM plans WHERE id = ?", id,
	).Scan(&p.ID, &p.Name, &p.VRAM, &p.Storage, &p.PriceHr, &p.PriceMonth, &p.MaxUsers, &active)
	if err != nil {
		return nil, err
	}
	p.Active = active == 1

	// Load pool
	pool, _ := DBListPool(id)
	p.Pool = pool
	return p, nil
}

func DBUpdatePlan(p DBPlan) error {
	_, err := db.Exec(
		"UPDATE plans SET name=?, vram=?, storage=?, price_hr=?, price_month=?, max_users=? WHERE id=?",
		p.Name, p.VRAM, p.Storage, p.PriceHr, p.PriceMonth, p.MaxUsers, p.ID,
	)
	return err
}

func DBDeletePlan(id string) error {
	_, err := db.Exec("DELETE FROM plans WHERE id = ?", id)
	return err
}

// --- Pool Operations ---

func DBAddPoolEntry(e DBPoolEntry) error {
	_, err := db.Exec(
		"INSERT INTO pool_entries (plan_id, priority, provider, label, machine_id, address, token) VALUES (?, ?, ?, ?, ?, ?, ?)",
		e.PlanID, e.Priority, e.Provider, e.Label, e.MachineID, e.Address, e.Token,
	)
	return err
}

func DBListPool(planID string) ([]DBPoolEntry, error) {
	rows, err := db.Query(
		"SELECT id, plan_id, priority, provider, label, machine_id, address, token FROM pool_entries WHERE plan_id = ? ORDER BY priority",
		planID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []DBPoolEntry
	for rows.Next() {
		var e DBPoolEntry
		rows.Scan(&e.ID, &e.PlanID, &e.Priority, &e.Provider, &e.Label, &e.MachineID, &e.Address, &e.Token)
		entries = append(entries, e)
	}
	return entries, nil
}

func DBDeletePoolEntry(id int) error {
	_, err := db.Exec("DELETE FROM pool_entries WHERE id = ?", id)
	return err
}

// --- Machine Operations ---

type DBMachine struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	GPU      string  `json:"gpu"`
	VRAM     string  `json:"vram"`
	PriceHr  float64 `json:"price_hr"`
	Status   string  `json:"status"`
	Address  string  `json:"address"`
	Provider       string  `json:"provider"`
	ProviderCostHr float64 `json:"provider_cost_hr"`
	RentedBy       string  `json:"rented_by,omitempty"`
}

func DBCreateMachine(m DBMachine) error {
	_, err := db.Exec(
		"INSERT INTO machines (id, name, gpu, vram, price_hr, status, address, provider, provider_cost_hr, rented_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		m.ID, m.Name, m.GPU, m.VRAM, m.PriceHr, m.Status, m.Address, m.Provider, m.ProviderCostHr, m.RentedBy,
	)
	return err
}

func DBListMachines() ([]DBMachine, error) {
	rows, err := db.Query("SELECT id, name, gpu, vram, price_hr, status, address, provider, provider_cost_hr, rented_by FROM machines ORDER BY price_hr")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var machines []DBMachine
	for rows.Next() {
		var m DBMachine
		rows.Scan(&m.ID, &m.Name, &m.GPU, &m.VRAM, &m.PriceHr, &m.Status, &m.Address, &m.Provider, &m.ProviderCostHr, &m.RentedBy)
		machines = append(machines, m)
	}
	return machines, nil
}

func DBDeleteMachine(id string) error {
	_, err := db.Exec("DELETE FROM machines WHERE id = ?", id)
	return err
}

// --- Transaction Log ---

func DBLogTransaction(userID int, amount float64, txType, description string) error {
	_, err := db.Exec(
		"INSERT INTO transactions (user_id, amount, type, description) VALUES (?, ?, ?, ?)",
		userID, amount, txType, description,
	)
	return err
}

func DBListTransactions(userID int) ([]map[string]interface{}, error) {
	rows, err := db.Query(
		"SELECT id, amount, type, description, created_at FROM transactions WHERE user_id = ? ORDER BY id DESC LIMIT 50",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []map[string]interface{}
	for rows.Next() {
		var id int
		var amount float64
		var txType, desc, createdAt string
		rows.Scan(&id, &amount, &txType, &desc, &createdAt)
		txs = append(txs, map[string]interface{}{
			"id": id, "amount": amount, "type": txType,
			"description": desc, "created_at": createdAt,
		})
	}
	return txs, nil
}

// --- Subscription Operations ---

func DBSubscribe(userID int, planID string) error {
	// Deactivate existing
	db.Exec("UPDATE subscriptions SET active = 0 WHERE user_id = ? AND active = 1", userID)

	_, err := db.Exec(
		"INSERT INTO subscriptions (user_id, plan_id) VALUES (?, ?)",
		userID, planID,
	)
	return err
}

func DBGetSubscription(userID int) (*DBPlan, error) {
	var planID string
	err := db.QueryRow(
		"SELECT plan_id FROM subscriptions WHERE user_id = ? AND active = 1", userID,
	).Scan(&planID)
	if err != nil {
		return nil, err
	}
	return DBGetPlan(planID)
}

func DBUnsubscribe(userID int) error {
	_, err := db.Exec("UPDATE subscriptions SET active = 0 WHERE user_id = ? AND active = 1", userID)
	return err
}

// --- User Machines ---

type DBUserMachine struct {
	ID       int    `json:"id"`
	UserID   int    `json:"user_id"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	Token    string `json:"token,omitempty"`
	Provider string `json:"provider"`
	Priority int    `json:"priority"`
}

func DBAddUserMachine(userID int, name, address, token, provider string, priority int) error {
	_, err := db.Exec(
		"INSERT INTO user_machines (user_id, name, address, token, provider, priority) VALUES (?, ?, ?, ?, ?, ?)",
		userID, name, address, token, provider, priority,
	)
	return err
}

func DBListUserMachines(userID int) ([]DBUserMachine, error) {
	rows, err := db.Query(
		"SELECT id, user_id, name, address, token, provider, priority FROM user_machines WHERE user_id = ? ORDER BY priority",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var machines []DBUserMachine
	for rows.Next() {
		var m DBUserMachine
		rows.Scan(&m.ID, &m.UserID, &m.Name, &m.Address, &m.Token, &m.Provider, &m.Priority)
		machines = append(machines, m)
	}
	return machines, nil
}

func DBDeleteUserMachine(userID, machineID int) error {
	_, err := db.Exec("DELETE FROM user_machines WHERE id = ? AND user_id = ?", machineID, userID)
	return err
}

// --- Job Operations ---

type DBJob struct {
	ID          string  `json:"id"`
	UserID      int     `json:"user_id"`
	PlanID      string  `json:"plan_id"`
	MachineUsed string  `json:"machine_used"`
	Status      string  `json:"status"`
	DurationS   float64 `json:"duration_s"`
	Cost        float64 `json:"cost"`
	CreatedAt   string  `json:"created_at"`
	FinishedAt  string  `json:"finished_at"`
}

func DBCreateJob(job DBJob) error {
	_, err := db.Exec(
		"INSERT INTO jobs (id, user_id, plan_id, machine_used, status) VALUES (?, ?, ?, ?, ?)",
		job.ID, job.UserID, job.PlanID, job.MachineUsed, job.Status,
	)
	return err
}

func DBListJobs(userID int) ([]DBJob, error) {
	rows, err := db.Query(
		"SELECT id, user_id, plan_id, machine_used, status, duration_s, cost, created_at, finished_at FROM jobs WHERE user_id = ? ORDER BY created_at DESC LIMIT 50",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []DBJob
	for rows.Next() {
		var j DBJob
		rows.Scan(&j.ID, &j.UserID, &j.PlanID, &j.MachineUsed, &j.Status, &j.DurationS, &j.Cost, &j.CreatedAt, &j.FinishedAt)
		jobs = append(jobs, j)
	}
	return jobs, nil
}

func DBUpdateJobStatus(jobID, status string) error {
	_, err := db.Exec("UPDATE jobs SET status = ? WHERE id = ?", status, jobID)
	return err
}

func DBGetUserByID(id int) (*DBUser, error) {
	u := &DBUser{}
	err := db.QueryRow(
		"SELECT id, email, token, credits, role, created_at FROM users WHERE id = ?", id,
	).Scan(&u.ID, &u.Email, &u.Token, &u.Credits, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func DBDeleteUser(id int) error {
	_, err := db.Exec("UPDATE users SET role = 'disabled' WHERE id = ?", id)
	return err
}

func DBUpdateMachine(m DBMachine) error {
	_, err := db.Exec(
		"UPDATE machines SET name=?, gpu=?, vram=?, price_hr=?, address=?, provider=? WHERE id=?",
		m.Name, m.GPU, m.VRAM, m.PriceHr, m.Address, m.Provider, m.ID,
	)
	return err
}

// --- Helpers ---

func nowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// --- Settings Operations ---
func DBSetSetting(key, value string) error {
	_, err := db.Exec("INSERT INTO kv_settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value=excluded.value", key, value)
	return err
}

func DBGetSetting(key string) string {
	var val string
	db.QueryRow("SELECT value FROM kv_settings WHERE key = ?", key).Scan(&val)
	return val
}
