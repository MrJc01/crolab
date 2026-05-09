package gateway

import (
	"context"
	"testing"
)

func TestHubRegistration(t *testing.T) {
	hub := NewHub()

	// Mock session com context e cancel mockados
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	
	session1 := &Session{ID: "session-1", ctx: ctx1, cancel: cancel1, isClosed: true} // isClosed: true previne o Close() de acessar socket nil
	session2 := &Session{ID: "session-2", ctx: ctx2, cancel: cancel2, isClosed: true}

	// Testa Registro
	hub.Register("session-1", session1)
	hub.Register("session-2", session2)

	if len(hub.sessions) != 2 {
		t.Errorf("Esperado 2 sessões registradas, obtido %d", len(hub.sessions))
	}

	// Testa Recuperação
	s, exists := hub.GetSession("session-1")
	if !exists || s.ID != "session-1" {
		t.Errorf("Falha ao recuperar session-1")
	}

	// Testa Remoção
	hub.Unregister("session-1")
	if len(hub.sessions) != 1 {
		t.Errorf("Esperado 1 sessão após remoção, obtido %d", len(hub.sessions))
	}

	_, exists = hub.GetSession("session-1")
	if exists {
		t.Errorf("A sessão ainda existe no mapa após Unregister")
	}
}
