#!/bin/bash
# ============================================
# Crolab E2E Playwright Test Runner + Report Generator
# Roda TODOS os cenários e gera relatório .md com prints e logs.
# ============================================
set -o pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
VENV_PYTHON="/tmp/crolab_screens/venv/bin/python"
REPORT_DIR="$ROOT_DIR/documentacao/relatorios"
REPORT_FILE="$REPORT_DIR/relatorio_e2e_playwright.md"
LOG_FILE="$SCRIPT_DIR/last_run.log"
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')

echo "════════════════════════════════════════"
echo "  🧪 CROLAB E2E PLAYWRIGHT SUITE"
echo "  $(date)"
echo "════════════════════════════════════════"

# 1. Compilar binário atualizado
echo "[1/4] 🔨 Compilando Crolab..."
cd "$ROOT_DIR"
go build -o crolab ./cmd/crolab/ 2>&1
echo "      ✅ Build OK"

# 2. Verificar Playwright
if [ ! -f "$VENV_PYTHON" ]; then
    echo "[2/4] 📦 Criando venv Playwright..."
    python3 -m venv /tmp/crolab_screens/venv
    $VENV_PYTHON -m pip install -q playwright pytest requests
    $VENV_PYTHON -m playwright install chromium
else
    echo "[2/4] ✅ Playwright venv encontrado"
fi

# 3. Executar testes com output capturado
echo "[3/4] 🚀 Rodando testes Playwright..."
echo ""

cd "$SCRIPT_DIR"
$VENV_PYTHON -m pytest \
    01_auth/ \
    02_dashboard/ \
    03_plans/ \
    04_machines/ \
    05_jobs/ \
    06_lab_colab/ \
    07_billing/ \
    08_admin/ \
    -v \
    --tb=long \
    -s \
    2>&1 | tee "$LOG_FILE"

EXIT_CODE=${PIPESTATUS[0]}

echo ""
echo "════════════════════════════════════════"
if [ $EXIT_CODE -eq 0 ]; then
    echo "  ✅ TODOS OS TESTES PASSARAM"
else
    echo "  ❌ HOUVE FALHAS (exit: $EXIT_CODE)"
fi
echo "════════════════════════════════════════"

# 4. Gerar relatório .md
echo "[4/4] 📝 Gerando relatório Markdown..."

TOTAL_TESTS=$(grep -c "PASSED\|FAILED" "$LOG_FILE" 2>/dev/null || echo "0")
PASSED=$(grep -c "PASSED" "$LOG_FILE" 2>/dev/null || echo "0")
FAILED=$(grep -c "FAILED" "$LOG_FILE" 2>/dev/null || echo "0")

mkdir -p "$REPORT_DIR"

cat > "$REPORT_FILE" << 'HEADER'
# Relatório E2E Playwright — Crolab Platform

> Gerado automaticamente pela suíte `tests/emular/run_all.sh`

HEADER

cat >> "$REPORT_FILE" << EOF
| Campo | Valor |
|---|---|
| **Data** | $TIMESTAMP |
| **Total de Testes** | $TOTAL_TESTS |
| **Aprovados** | ✅ $PASSED |
| **Falhas** | ❌ $FAILED |
| **Duração** | $(tail -1 "$LOG_FILE" | grep -oP '\d+\.\d+s' || echo "N/A") |
| **Binário** | $(./crolab version 2>/dev/null || echo "N/A") |

---

## Módulos Testados

### 01 — Autenticação (Auth)
Verifica Landing Page, Login SSO, Registro de novo usuário.

EOF

# Função para adicionar screenshots de um módulo
add_screenshots() {
    local module_dir="$1"
    local module_name="$2"
    local screenshot_dir="$SCRIPT_DIR/$module_dir/screenshots"

    if [ -d "$screenshot_dir" ]; then
        for img in "$screenshot_dir"/*.png; do
            if [ -f "$img" ]; then
                local basename=$(basename "$img")
                local rel_path="../../tests/emular/$module_dir/screenshots/$basename"
                echo "![${basename%.png}]($rel_path)" >> "$REPORT_FILE"
                echo "" >> "$REPORT_FILE"
            fi
        done
    fi
}

add_screenshots "01_auth" "Auth"

cat >> "$REPORT_FILE" << 'EOF'
### 02 — Dashboard (Home)
Verifica métricas de créditos, plano ativo, máquinas, e integração terminal.

EOF
add_screenshots "02_dashboard" "Dashboard"

cat >> "$REPORT_FILE" << 'EOF'
### 03 — Planos e Assinaturas
Verifica catálogo de planos GPU, cards de preço, e fluxo de assinatura.

EOF
add_screenshots "03_plans" "Plans"

cat >> "$REPORT_FILE" << 'EOF'
### 04 — Máquinas P2P (Nodos GPU)
Verifica grid de GPUs disponíveis, formulário Bridge RPC, e botões de aluguel.

EOF
add_screenshots "04_machines" "Machines"

cat >> "$REPORT_FILE" << 'EOF'
### 05 — Fila de Jobs (Execuções)
Verifica tabela de jobs, colunas da fila, e botão de sincronização.

EOF
add_screenshots "05_jobs" "Jobs"

cat >> "$REPORT_FILE" << 'EOF'
### 06 — Editor Colab-Style (Monaco + WebSocket Kernel)
Verifica que o Monaco Editor renderiza nativamente (sem iframe), que o WebSocket Kernel executa Python, e que os logs aparecem em tempo real.

EOF
add_screenshots "06_lab_colab" "Lab Colab"

cat >> "$REPORT_FILE" << 'EOF'
### 07 — Financeiro (Billing)
Verifica saldo de créditos, compra via botões Mint, e ledger de transações.

EOF
add_screenshots "07_billing" "Billing"

cat >> "$REPORT_FILE" << 'EOF'
### 08 — Painel Admin
Verifica navegação completa entre Dashboard, Planos, Máquinas, Usuários e Logs.

EOF
add_screenshots "08_admin" "Admin"

# Adicionar log completo ao final
cat >> "$REPORT_FILE" << EOF

---

## Log Completo da Execução

\`\`\`
$(cat "$LOG_FILE")
\`\`\`

---

## Inventário de Screenshots

| Arquivo | Módulo |
|---|---|
EOF

find "$SCRIPT_DIR" -name "*.png" -type f | sort | while read f; do
    rel=$(echo "$f" | sed "s|$SCRIPT_DIR/||")
    module=$(echo "$rel" | cut -d'/' -f1)
    echo "| \`$rel\` | $module |" >> "$REPORT_FILE"
done

echo "" >> "$REPORT_FILE"
echo "> Relatório gerado em: $TIMESTAMP" >> "$REPORT_FILE"

echo ""
echo "📄 Relatório salvo em: $REPORT_FILE"
echo ""

# Listar screenshots
echo "📸 Screenshots gerados:"
find "$SCRIPT_DIR" -name "*.png" -type f | sort
echo ""

exit $EXIT_CODE
