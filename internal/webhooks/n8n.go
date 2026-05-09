package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// N8nClient gerencia disparos assíncronos de eventos para uma instância n8n
type N8nClient struct {
	WebhookURL string
	HTTPClient *http.Client
}

// NewN8nClient inicializa o despachante de webhooks
func NewN8nClient(url string) *N8nClient {
	return &N8nClient{
		WebhookURL: url,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// NotifyNewTenant dispara um payload JSON informando um novo cadastro
// Utiliza goroutine internamente para não bloquear o fluxo da API do Gateway
func (n *N8nClient) NotifyNewTenant(tenantID string, email string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		payload := map[string]string{
			"event":     "tenant_created",
			"tenant_id": tenantID,
			"email":     email,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}

		body, _ := json.Marshal(payload)
		req, err := http.NewRequestWithContext(ctx, "POST", n.WebhookURL, bytes.NewBuffer(body))
		if err != nil {
			log.Printf("[N8n Webhook] Erro ao criar request: %v", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := n.HTTPClient.Do(req)
		if err != nil {
			log.Printf("[N8n Webhook] Falha ao despachar notificação para %s: %v", email, err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 300 {
			log.Printf("[N8n Webhook] n8n retornou erro HTTP %d", resp.StatusCode)
		}
	}()
}
