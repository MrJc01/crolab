#!/bin/bash
echo "🧠 Iniciando Kernel ZeroMQ (Mock) na porta 5555..."
cd scratch/poc_kernel || { echo "Pasta do Kernel não encontrada!"; exit 1; }
source venv/bin/activate
python3 kernel.py
