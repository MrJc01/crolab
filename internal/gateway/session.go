package gateway

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/go-zeromq/zmq4"
	"github.com/gorilla/websocket"
)

// Session representa uma ponte bidirecional entre um cliente WebSocket e um Kernel ZeroMQ
type Session struct {
	ID        string
	WSConn    *websocket.Conn
	ZMQSocket zmq4.Socket
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.Mutex
	isClosed  bool
}

// NewSession inicializa a ponte. Espera que o ZMQ endpoint já esteja ativo (ex: "tcp://127.0.0.1:5555")
func NewSession(id string, ws *websocket.Conn, zmqEndpoint string) (*Session, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Cria o socket REQ do ZeroMQ para mandar e receber código
	zmqReq := zmq4.NewReq(ctx)
	err := zmqReq.Dial(zmqEndpoint)
	if err != nil {
		cancel()
		return nil, err
	}

	return &Session{
		ID:        id,
		WSConn:    ws,
		ZMQSocket: zmqReq,
		ctx:       ctx,
		cancel:    cancel,
		isClosed:  false,
	}, nil
}

// Start inícia o loop de escuta e roteamento de mensagens
func (s *Session) Start() {
	// A goroutine de leitura do WebSocket (Frontend -> ZeroMQ)
	go func() {
		defer s.Close()
		for {
			_, msg, err := s.WSConn.ReadMessage()
			if err != nil {
				log.Printf("[Session %s] WS Read Error: %v", s.ID, err)
				break
			}

			// Recebeu código do WS, manda para o Kernel ZMQ
			err = s.ZMQSocket.Send(zmq4.NewMsg(msg))
			if err != nil {
				log.Printf("[Session %s] ZMQ Send Error: %v", s.ID, err)
				break
			}

			// Aguarda resposta do Kernel
			reply, err := s.ZMQSocket.Recv()
			if err != nil {
				log.Printf("[Session %s] ZMQ Recv Error: %v", s.ID, err)
				break
			}

			// Envia a resposta de volta ao Frontend
			s.mu.Lock()
			err = s.WSConn.WriteMessage(websocket.TextMessage, reply.Frames[0])
			s.mu.Unlock()
			if err != nil {
				log.Printf("[Session %s] WS Write Error: %v", s.ID, err)
				break
			}
		}
	}()
}

// Close encerra as conexões graciosamente
func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.isClosed {
		return
	}
	s.isClosed = true
	
	s.cancel()               // Cancela o contexto do ZMQ
	s.ZMQSocket.Close()      // Fecha o socket
	s.WSConn.Close()         // Fecha o websocket
	
	log.Printf("[Session %s] Conexões encerradas.", s.ID)
}
