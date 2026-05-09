#!/bin/bash
echo "🧪 Iniciando Bateria de Testes (TDD/E2E)..."
echo "-----------------------------------------"

# Testa os pacotes de backend (Go)
echo "[Go] Testando Gateway..."
go test -v ./internal/gateway

echo "-----------------------------------------"
echo "[Go] Testando Storage S3..."
go test -v ./internal/storage

echo "-----------------------------------------"
echo "✅ Testes Backend Concluídos."

# Testa o frontend (Se houver testes configurados no package.json como jest/vitest)
# cd web/frontend && npm test
