package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"strings"

	"github.com/crolab/core/internal/cloud"
)

func TestAuthFuzzing(t *testing.T) {
	cloud.InitDB(":memory:")
	
	mux := cloud.BuildMux("./web")

	// 1. Tentar SQL Injection no Login Body ("admin' OR '1'='1")
	reqBody1, _ := json.Marshal(map[string]string{
		"email":    "admin' OR '1'='1",
		"password": "senha",
	})
	req1 := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(reqBody1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	mux.ServeHTTP(w1, req1)

	if w1.Result().StatusCode == 200 {
		t.Errorf("Vulnerabilidade Crítica! SQL Injection foi interpretado e permitiu login: %s", w1.Body.String())
	}
	if !strings.Contains(w1.Body.String(), "credenciais inválidas") {
		t.Errorf("Comportamento não tratado no DB ao tentar injeção: %s", w1.Body.String())
	}

	// 2. Fuzzing JSON quebrado ou malicioso gigante 
	payloadSize := make([]byte, 5*1024*1024) // 5MB JSON
	for i := range payloadSize {
		payloadSize[i] = 'A'
	}
	// json syntax error malicioso
	req2 := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(payloadSize))
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, req2)

	// Pode rejeitar no Decoder com 400 ou panicar se não tratado. Como é em Go padrão, json.Decoder engole e dá erro 400.
	if w2.Result().StatusCode == 200 || w2.Result().StatusCode == 500 {
		t.Errorf("Fuzzing gigante causou falha inesperada: status %d", w2.Result().StatusCode)
	}

	// 3. Register Injection Script Tag (XSS nas chaves primary)
	reqBody3, _ := json.Marshal(map[string]string{
		"email":    "<script>alert(1)</script>@hack.com",
		"password": "senha",
	})
	req3 := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(reqBody3))
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	mux.ServeHTTP(w3, req3)

	// Deveria registrar sanitizado ou dar 201 normal. Vamos testar /auth/login e ler se escapa HTML
	if w3.Result().StatusCode != 201 {
		t.Errorf("Registro falhou: %s", w3.Body.String())
	}
}
