#!/bin/bash
# Crolab SRE Deployment Script - v1.0
# curl -sSL https://crolab.crom.run/install | bash

set -e

echo "=========================================="
echo "⚡ Crolab P2P Orchestrator Installer ⚡"
echo "=========================================="

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Arquitetura não suportada: $ARCH"; exit 1 ;;
esac

echo "Baixando release Crolab para $OS-$ARCH..."
DOWNLOAD_URL="https://github.com/crolab/core/releases/latest/download/crolab-${OS}-${ARCH}"
# wget/curl simulado aqui
curl -fsSL -o /tmp/crolab "$DOWNLOAD_URL" || { echo "Falha ao baixar do GitHub releases"; exit 1; }

chmod +x /tmp/crolab
sudo mv /tmp/crolab /usr/local/bin/crolab

echo "✅ Crolab instalado com sucesso em /usr/local/bin"
echo "   Inicie o Provider Node com: crolab provider start"
