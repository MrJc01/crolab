package gateway

import (
	"log"
	"sync"
)

// Hub mantém o registro global das sessões ativas (Frontend WS <-> Go <-> ZeroMQ)
type Hub struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewHub cria e retorna um novo Hub de conexões
func NewHub() *Hub {
	return &Hub{
		sessions: make(map[string]*Session),
	}
}

// Register adiciona uma nova sessão ao Hub
func (h *Hub) Register(sessionID string, s *Session) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sessions[sessionID] = s
	log.Printf("[Hub] Sessão %s registrada. Total: %d", sessionID, len(h.sessions))
}

// Unregister remove uma sessão do Hub e garante seu fechamento
func (h *Hub) Unregister(sessionID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if s, exists := h.sessions[sessionID]; exists {
		s.Close()
		delete(h.sessions, sessionID)
		log.Printf("[Hub] Sessão %s removida. Total: %d", sessionID, len(h.sessions))
	}
}

// GetSession recupera uma sessão ativa pelo ID
func (h *Hub) GetSession(sessionID string) (*Session, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	s, exists := h.sessions[sessionID]
	return s, exists
}

// CloseAll encerra todas as sessões ativas durante o shutdown do servidor
func (h *Hub) CloseAll() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for id, s := range h.sessions {
		s.Close()
		delete(h.sessions, id)
	}
	log.Println("[Hub] Todas as sessões foram encerradas.")
}
