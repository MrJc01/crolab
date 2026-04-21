package db_test

import (
	"os"
	"testing"
	"github.com/crolab/core/internal/cloud"
)

func setupTestDB(t *testing.T) string {
	dbPath := "/tmp/crolab_test_crud.db"
	_ = os.Remove(dbPath)
	
	if err := cloud.InitDB(dbPath); err != nil {
		t.Fatalf("Falha InitDB: %v", err)
	}
	return dbPath
}

func TestMigrationIdempotency(t *testing.T) {
	dbPath := setupTestDB(t)
	defer os.Remove(dbPath)

	// Idempotência: Chamar InitDB na mesma base migrada não pode falhar ou resetar os dados
	err := cloud.InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB falhou na segunda vez: %v", err)
	}
}

func TestUserCRUD(t *testing.T) {
	dbPath := setupTestDB(t)
	defer os.Remove(dbPath)

	// Create
	rawPass := "senha123"
	_, err := cloud.DBCreateUser("crud@test.com", rawPass, "client", "10.0.0.1")
	if err != nil {
		t.Fatalf("Falhou ao criar user: %v", err)
	}

	// Read
	u, err := cloud.DBGetUserByEmail("crud@test.com")
	if err != nil || u == nil {
		t.Fatalf("Falhou ao buscar user: %v", err)
	}

	hashPass, _ := cloud.DBGetPasswordHash("crud@test.com")
	if hashPass == rawPass {
		t.Fatalf("CRÍTICO: Hash não mascarado, exposto plain text: %s", hashPass)
	}

	// Update (Credits)
	_, _ = cloud.DBUpdateCredits(u.ID, 120.5)
	u2, _ := cloud.DBGetUserByEmail("crud@test.com")
	if u2.Credits != 120.5 {
		t.Errorf("Esperava 120.5 creditos, obteve: %f", u2.Credits)
	}

	// Delete (soft-delete: role → 'disabled')
	_ = cloud.DBDeleteUser(u.ID)
	u3, err := cloud.DBGetUserByEmail("crud@test.com")
	if err != nil || u3 == nil {
		t.Fatalf("Soft-delete falhou ao buscar user: %v", err)
	}
	if u3.Role != "disabled" {
		t.Fatalf("Esperava role='disabled' após soft-delete, obteve: %s", u3.Role)
	}
}

func TestPlanMachineCRUD(t *testing.T) {
	dbPath := setupTestDB(t)
	defer os.Remove(dbPath)

	p := cloud.DBPlan{ID: "plano_crud", Name: "Plano CR", VRAM: "10GB", Storage: "10GB", PriceHr: 1.0, MaxUsers: 5}
	err := cloud.DBCreatePlan(p)
	if err != nil {
		t.Fatalf("DBCreatePlan: %v", err)
	}

	m := cloud.DBMachine{ID: "mach_crud", Name: "Nvidia TX", PriceHr: 0.5, ProviderCostHr: 0.2, Provider: "testp"}
	err = cloud.DBCreateMachine(m)
	if err != nil {
		t.Fatalf("DBCreateMachine: %v", err)
	}

	cloud.DBUpdatePlan(cloud.DBPlan{ID: "plano_crud", Name: "Plano Editado"})
	p2, _ := cloud.DBGetPlan("plano_crud")
	if p2.Name != "Plano Editado" {
		t.Errorf("Esperava 'Plano Editado', obteve %s", p2.Name)
	}
	
	cloud.DBDeletePlan("plano_crud")
	p3, _ := cloud.DBGetPlan("plano_crud")
	if p3 != nil {
		t.Errorf("Plano ainda existia.")
	}
}
