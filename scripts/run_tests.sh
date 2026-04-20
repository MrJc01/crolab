#!/usr/bin/env bash
#
# Crolab — Script de Testes Completo com Relatório
# Uso: ./scripts/run_tests.sh
#
# Gera relatório em tests/reports/report_YYYYMMDD_HHMMSS.txt
#

set -uo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPORT_DIR="tests/reports"
REPORT="$REPORT_DIR/report_${TIMESTAMP}.txt"
mkdir -p "$REPORT_DIR"

TOTAL=0
PASSED=0
FAILED=0
ERRORS=""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

header() {
    echo ""
    echo "═══════════════════════════════════════════════════════════"
    echo "  $1"
    echo "═══════════════════════════════════════════════════════════"
}

log() {
    echo "$1" | tee -a "$REPORT"
}

run_suite() {
    local name="$1"
    local pkg="$2"
    local extra="${3:-}"

    TOTAL=$((TOTAL + 1))
    echo -ne "  ${BLUE}▸${NC} $name... "
    log "" >> /dev/null
    log "━━━ $name ━━━" >> /dev/null

    OUTPUT=$(go test -v -count=1 $extra "$pkg" 2>&1)
    EXIT=$?

    echo "$OUTPUT" >> "$REPORT"
    echo "" >> "$REPORT"

    if [ $EXIT -eq 0 ]; then
        PASSED=$((PASSED + 1))
        # Count individual tests
        TEST_COUNT=$(echo "$OUTPUT" | grep -c "^--- PASS" || true)
        DURATION=$(echo "$OUTPUT" | grep "^ok" | awk '{print $NF}' || echo "?")
        echo -e "${GREEN}✓ PASS${NC} ($TEST_COUNT testes, $DURATION)"
        log "RESULTADO: PASS ($TEST_COUNT testes, $DURATION)"
    else
        FAILED=$((FAILED + 1))
        FAIL_TESTS=$(echo "$OUTPUT" | grep "^--- FAIL" | sed 's/--- FAIL: /  ✗ /' || echo "  (unknown)")
        echo -e "${RED}✗ FAIL${NC}"
        echo "$FAIL_TESTS"
        ERRORS="$ERRORS\n[$name]\n$FAIL_TESTS\n"
        log "RESULTADO: FAIL"
    fi
}

# ===========================================================

header "CROLAB — SUÍTE COMPLETA DE TESTES"
echo ""
echo "  Data:     $(date '+%Y-%m-%d %H:%M:%S')"
echo "  Go:       $(go version | awk '{print $3}')"
echo "  OS:       $(uname -s)/$(uname -m)"
echo "  Relatório: $REPORT"
echo ""

log "CROLAB TEST REPORT — $TIMESTAMP" >> "$REPORT"
log "Go: $(go version)" >> "$REPORT"
log "OS: $(uname -s)/$(uname -m)" >> "$REPORT"
log "======================================" >> "$REPORT"
log "" >> "$REPORT"

# --- Build ---
header "FASE 0: BUILD"
echo -ne "  ${BLUE}▸${NC} Compilando binário... "
BUILD_OUT=$(go build -o crolab ./cmd/crolab/ 2>&1)
if [ $? -eq 0 ]; then
    SIZE=$(du -sh crolab | awk '{print $1}')
    echo -e "${GREEN}✓ OK${NC} ($SIZE)"
    log "BUILD: OK ($SIZE)" >> "$REPORT"
else
    echo -e "${RED}✗ FALHOU${NC}"
    echo "$BUILD_OUT"
    log "BUILD: FALHOU" >> "$REPORT"
    log "$BUILD_OUT" >> "$REPORT"
    echo ""
    echo -e "${RED}Build falhou. Abortando testes.${NC}"
    exit 1
fi

# --- Unit Tests ---
header "FASE 1: TESTES UNITÁRIOS"
run_suite "ZipDir (compressão)"              "./tests/unit/..."
run_suite "Zip/Unzip (roundtrip + edge)"     "./tests/zip/..."
run_suite "Config CRUD (add/rm/sort/hash)"   "./tests/config/..."

# --- Integration Tests ---
header "FASE 2: TESTES DE INTEGRAÇÃO"
run_suite "gRPC (submit/auth/timeout)"       "./tests/grpc/..."
run_suite "CLI Completa (help/status/crud)"  "./tests/cli/..."

# --- Load Tests ---
header "FASE 3: TESTES DE CARGA"
run_suite "Stress 50x gRPC (chaos loop)"     "./tests/load/..."

# --- Cloud API Tests ---
header "FASE 4: CLOUD API (Auth + Billing + Machines + Admin + Client)"
run_suite "Cloud API (22 testes)"  "./tests/cloud/..." "-timeout 30s"

# --- E2E Tests ---
header "FASE 5: E2E (Fluxo Completo Admin→Client)"
run_suite "E2E Full Flow (15 steps)"  "./tests/e2e/..." "-timeout 30s"

# ===========================================================

header "RELATÓRIO FINAL"
echo ""
PERCENT=$((PASSED * 100 / TOTAL))

echo "  ┌─────────────────────────────────┐"
echo "  │  Suítes executadas:  $TOTAL              │"
if [ $FAILED -eq 0 ]; then
    echo -e "  │  ${GREEN}Passaram:            $PASSED${NC}              │"
    echo "  │  Falharam:            0              │"
else
    echo "  │  Passaram:            $PASSED              │"
    echo -e "  │  ${RED}Falharam:            $FAILED${NC}              │"
fi
echo "  │  Taxa de sucesso:    ${PERCENT}%            │"
echo "  └─────────────────────────────────┘"

log "" >> "$REPORT"
log "======================================" >> "$REPORT"
log "RESUMO" >> "$REPORT"
log "  Suítes:  $TOTAL" >> "$REPORT"
log "  Passed:  $PASSED" >> "$REPORT"
log "  Failed:  $FAILED" >> "$REPORT"
log "  Rate:    ${PERCENT}%" >> "$REPORT"

if [ $FAILED -gt 0 ]; then
    echo ""
    echo -e "  ${RED}Falhas:${NC}"
    echo -e "$ERRORS"
    log "FALHAS:" >> "$REPORT"
    echo -e "$ERRORS" >> "$REPORT"
fi

echo ""
echo "  📄 Relatório completo: $REPORT"
echo ""

# Count total individual tests from report
INDIVIDUAL_TESTS=$(grep -c "^--- PASS\|^--- FAIL" "$REPORT" 2>/dev/null || echo "?")
echo "  Total de testes individuais: $INDIVIDUAL_TESTS"
echo ""

log "Total testes individuais: $INDIVIDUAL_TESTS" >> "$REPORT"

exit $FAILED
