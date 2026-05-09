package main

import (
	"log"
	"net/http"
	"github.com/crolab/core/internal/gateway"
	"github.com/crolab/core/internal/web"
)

func main() {
	log.Println("⚡ Crolab V3 Master - Inicializando Servidor de Testes")

	// 1. Inicializa o Hub de WebSockets (Fase 2)
	hub := gateway.NewHub()
	defer hub.CloseAll()

	// 2. Registra o Endpoint do Gateway (Fase 2)
	http.HandleFunc("/ws", gateway.WSHandler(hub))

	// 3. Registra o Endpoint de Métricas do Prometheus (Fase 6)
	http.Handle("/metrics", gateway.MetricsHandler())

	// 4. Registra o Frontend React Embutido (Fases 3 e 7)
	// Como a pasta "dist" não existe até compilar, para desenvolvimento
	// vamos sugerir rodar o React no npm e o Go separadamente, mas 
	// deixaremos a rota mapeada (ignorando erro se dist não existir).
	http.Handle("/", web.FrontendHandler())

	log.Println("✅ Servidor rodando na porta :8080")
	log.Println("   - Frontend (se compilado): http://localhost:8080/")
	log.Println("   - Gateway WS: ws://localhost:8080/ws")
	log.Println("   - Métricas: http://localhost:8080/metrics")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Erro fatal no servidor: %v", err)
	}
}
