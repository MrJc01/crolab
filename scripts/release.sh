#!/bin/bash
# scripts/release.sh - Build via Makefile e faz Push de tag para Github Release

set -e

VERSION="0.2.0"

echo "🔨 Compilando todos os binários cross-platform..."
make build-all

echo "📦 Preparando tag semântica v${VERSION}..."
git tag -a "v${VERSION}" -m "Release v${VERSION} - Orchestration Ready" || true
git push origin "v${VERSION}" || true

echo "🚢 Lançando Release v${VERSION} (requer gh cli)..."
gh release create "v${VERSION}" ./bin/crolab-* --title "Crolab v${VERSION}" --notes "Crolab Single Binary P2P Engine"

echo "✅ Release criado com sucesso!"
