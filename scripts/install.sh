#!/bin/bash
# scripts/install.sh - Script público de instalação do Crolab

set -e

echo "🚀 Iniciando instalação do Crolab (P2P GPU Orchestrator)..."

OS=$(uname -s | tr A-Z a-z)
ARCH=$(uname -m)

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "arm64" ] || [ "$ARCH" = "aarch64" ]; then
    ARCH="arm64"
fi

VERSION="0.2.0"
BINARY_URL="https://github.com/crolab/core/releases/download/v${VERSION}/crolab-${OS}-${ARCH}"

echo "⬇️  Baixando Crolab v${VERSION} para ${OS}/${ARCH}..."
curl -sL $BINARY_URL -o /tmp/crolab

echo "📦 Instalando em /usr/local/bin..."
sudo mv /tmp/crolab /usr/local/bin/crolab
sudo chmod +x /usr/local/bin/crolab

echo "✅ Sucesso! O Crolab foi instalado."
echo "➡️  Rode 'crolab help' para começar."
