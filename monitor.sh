#!/bin/bash

# Cores para o terminal
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Função para matar subprocessos no exit
cleanup() {
    echo -e "\n${RED}🛑 Encerrando todos os processos do Crolab...${NC}"
    # Mata os processos iniciados pelos filhos
    pkill -f "python3 kernel.py"
    pkill -f "cmd/crolab-v3/main.go"
    pkill -f "npm run dev"
    exit 0
}

# Captura Ctrl+C
trap cleanup SIGINT

show_menu() {
    clear
    echo -e "${BLUE}==========================================${NC}"
    echo -e "${GREEN}      ⚡ Crolab Dev Monitor V3 ⚡         ${NC}"
    echo -e "${BLUE}==========================================${NC}"
    echo "1. 🟢 Iniciar Todo o Ambiente Dev (Backend + Frontend + Kernel)"
    echo "2. 🚀 Iniciar apenas o Frontend (React SPA)"
    echo "3. ⚙️ Iniciar apenas o Backend (Go Gateway)"
    echo "4. 🧠 Iniciar apenas o Kernel (ZeroMQ Mock Python)"
    echo "5. 🧪 Rodar Todos os Testes Unitários"
    echo "6. ❌ Sair"
    echo -e "${BLUE}==========================================${NC}"
    echo -n "Escolha uma opção [1-6]: "
}

start_all() {
    echo -e "${YELLOW}Iniciando tudo em background. Pressione [Ctrl+C] para derrubar todos.${NC}"
    
    # Roda o Kernel
    ./scripts/dev/start-kernel.sh &
    KERNEL_PID=$!
    
    # Roda o Backend
    ./scripts/dev/start-backend.sh &
    BACKEND_PID=$!
    
    # Roda o Frontend
    ./scripts/dev/start-frontend.sh &
    FRONTEND_PID=$!
    
    # Aguarda processos (Isso fará o script segurar o terminal e os logs)
    wait $KERNEL_PID $BACKEND_PID $FRONTEND_PID
}

while true; do
    show_menu
    read option
    case $option in
        1)
            start_all
            ;;
        2)
            ./scripts/dev/start-frontend.sh
            read -p "Pressione enter para voltar ao menu..."
            ;;
        3)
            ./scripts/dev/start-backend.sh
            read -p "Pressione enter para voltar ao menu..."
            ;;
        4)
            ./scripts/dev/start-kernel.sh
            read -p "Pressione enter para voltar ao menu..."
            ;;
        5)
            ./scripts/dev/run-tests.sh
            read -p "Pressione enter para voltar ao menu..."
            ;;
        6)
            echo "Saindo..."
            exit 0
            ;;
        *)
            echo -e "${RED}Opção inválida!${NC}"
            sleep 1
            ;;
    esac
done
