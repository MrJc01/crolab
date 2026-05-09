#!/bin/bash
echo "🚀 Iniciando Frontend (React/Vite)..."
cd web/frontend || { echo "Pasta web/frontend não encontrada!"; exit 1; }
npm run dev
