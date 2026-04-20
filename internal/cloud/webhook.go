package cloud

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func SendWebhook(eventType string, data map[string]interface{}) {
	whUrl := os.Getenv("CROLAB_WEBHOOK_URL")
	if whUrl == "" {
		return
	}

	payload := map[string]interface{}{
		"event":     eventType,
		"timestamp": time.Now().Format(time.RFC3339),
		"data":      data,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Falha ao preparar webhook JSON", "err", err)
		return
	}

	go func() {
		client := &http.Client{Timeout: 5 * time.Second}
		req, _ := http.NewRequest("POST", whUrl, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Crolab-SRE-Webhook/1.0")

		resp, err := client.Do(req)
		if err != nil {
			slog.Warn("Falha ao entregar webhook SRE", "url", whUrl, "err", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			slog.Warn("Webhook SRE respondeu com erro", "status", resp.StatusCode)
		}
	}()
}
