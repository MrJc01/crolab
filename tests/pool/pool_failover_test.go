package pool_test

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Mock Node API
type MockNode struct {
	ID        string
	WillFail  bool
	WillDelay bool
	Address   string
}

func (m *MockNode) RunJob(ctx context.Context, jobID string) error {
	if m.WillDelay {
		select {
		case <-time.After(3 * time.Second):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if m.WillFail {
		return errors.New("connection_refused")
	}
	return nil
}

// Cascata Router: Itera sobre nodes baseados na prioridade (Testable Unit)
func RouteJobCascata(ctx context.Context, jobID string, nodes []MockNode) (string, error) {
	var errs []error
	for _, n := range nodes {
		nodeCtx, cancel := context.WithTimeout(ctx, 2*time.Second) // SRE: Timeout curto para failover
		err := n.RunJob(nodeCtx, jobID)
		cancel()

		if err == nil {
			return n.ID, nil // Sucesso no primeiro que rodou!
		}
		errs = append(errs, err)
	}
	return "", errors.New("todas_instancias_offline_timeout")
}

func TestCascataFailover(t *testing.T) {
	nodes := []MockNode{
		{ID: "node1_high_priority", WillFail: true},
		{ID: "node2_low_priority", WillFail: false},
	}

	winner, err := RouteJobCascata(context.Background(), "job-abc", nodes)
	if err != nil {
		t.Fatalf("Esperava fallback sucesso no node2, mas falhou tudo: %v", err)
	}

	if winner != "node2_low_priority" {
		t.Errorf("Esperava node2 ser o escalonado, obteve %s", winner)
	}
}

func TestCascataAllOffline(t *testing.T) {
	nodes := []MockNode{
		{ID: "node1", WillFail: true},
		{ID: "node2", WillFail: true},
	}

	_, err := RouteJobCascata(context.Background(), "job-123", nodes)
	if err == nil || err.Error() != "todas_instancias_offline_timeout" {
		t.Errorf("Esperava falha total de fila (todas_instancias_offline_timeout), obteve %v", err)
	}
}

func TestCascataTimeout(t *testing.T) {
	nodes := []MockNode{
		{ID: "node1_slow", WillDelay: true},
		{ID: "node2_fast", WillFail: false},
	}

	winner, err := RouteJobCascata(context.Background(), "job-slow", nodes)
	if err != nil {
		t.Fatalf("Cascata abortada incorretamente. Timeout falhou: %v", err)
	}

	if winner != "node2_fast" {
		t.Errorf("Node 1 demorou, timeout devia despachar para Node 2. Obtido: %s", winner)
	}
}
