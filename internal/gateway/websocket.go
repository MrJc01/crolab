package gateway

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Permite conexões de qualquer origem para facilitar o desenvolvimento local
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WSHandler é o handler HTTP que faz o upgrade da conexão para WebSocket
func WSHandler(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
			return
		}

		// Na Fase 2 de testes, estamos assumindo que cada conexão pede um ID de Kernel
		// Em produção, isso virá de um JWT ou query param.
		sessionID := r.URL.Query().Get("id")
		if sessionID == "" {
			conn.Close()
			return
		}

		// Endereço mockado do Kernel (criado na Fase 1 da PoC ZMQ)
		// Em produção isso será gerido dinamicamente pelo Firecracker
		zmqEndpoint := fmt.Sprintf("tcp://127.0.0.1:5555")

		session, err := NewSession(sessionID, conn, zmqEndpoint)
		if err != nil {
			conn.Close()
			return
		}

		// Registra e inicia a sessão
		hub.Register(sessionID, session)
		session.Start()

		// Tratamento rudimentar para garantir limpeza se o cliente desconectar imediatamente (mock ping loop)
		// Na vida real, teremos pings e pongs. O fechamento é feito pelo ReadError dentro de Start()
	}
}
