"""
07 — Validação do Stateful Kernel (Jupyter Clone)
Testes unitários e funcionais da Fase 2 (Python Engine, Toggles e Persistência).

Design: Testes 1-5 (UI pura) verificam DOM — NÃO precisam do Monaco.
        Testes 6-10 (Kernel) compartilham UMA sessão de browser.
"""
import os
import pytest
from playwright.sync_api import sync_playwright

SCREENSHOTS_DIR = os.path.join(os.path.dirname(__file__), "screenshots")
os.makedirs(SCREENSHOTS_DIR, exist_ok=True)


def _login_and_goto_lab(page, client_url, wait_monaco=False):
    """Helper para logar e navegar ao Lab.
    
    Args:
        wait_monaco: Se True, espera o Monaco CDN carregar (lento, ~10s).
                     Para testes de UI pura, não é necessário.
    """
    page.goto(client_url)
    page.wait_for_load_state("networkidle")

    if page.locator("#btn-start-landing").is_visible():
        page.click("#btn-start-landing")

    page.wait_for_selector("#auth-email", state="visible", timeout=5000)
    page.fill("#auth-email", "admin@crolab.com")
    page.fill("#auth-password", "admin123")
    page.click("#btn-auth-login")

    page.wait_for_selector("#home-section", state="visible", timeout=5000)

    page.click("#btn-tab-lab")
    page.wait_for_selector("#lab-section", state="visible", timeout=5000)
    
    if wait_monaco:
        page.wait_for_function("window.monacoEditor !== undefined", timeout=30000)


def _inject_and_run(page, code, wait_ms=3000):
    """Inject code into Monaco and run it, returning the output text."""
    page.evaluate(f"window.monacoEditor.setValue({repr(code)});")
    page.wait_for_timeout(300)
    page.click("#btn-run-cell")
    page.wait_for_timeout(wait_ms)
    return page.locator("#lab-output").inner_text()


# ═══════════════════════════════════════════════════════════════
#  TESTES 1-5: UI PURA (não precisam do Monaco)
# ═══════════════════════════════════════════════════════════════

def test_lab_url_hash_routing(client_url, crolab_server):
    """(1) Redirecionamento automático via URL Hash #lab."""
    with sync_playwright() as p:
        b = p.chromium.launch(headless=True)
        page = b.new_page(viewport={"width": 1440, "height": 900})

        page.goto(client_url + "#lab")
        page.wait_for_load_state("networkidle")
        if page.locator("#btn-start-landing").is_visible():
            page.click("#btn-start-landing")

        page.wait_for_selector("#auth-email", state="visible", timeout=5000)
        page.fill("#auth-email", "admin@crolab.com")
        page.fill("#auth-password", "admin123")
        page.click("#btn-auth-login")

        page.wait_for_selector("#lab-section", state="visible", timeout=8000)

        assert page.locator("#lab-section").is_visible(), "Hash routing falhou"
        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "01_hash_routing.png"))
        print("[PASS] Hash routing #lab funciona")
        b.close()


def test_lab_sidebar_toggle(client_url, crolab_server):
    """(2) Testar Hide/Show do Sidebar Global."""
    with sync_playwright() as p:
        b = p.chromium.launch(headless=True)
        page = b.new_page()
        _login_and_goto_lab(page, client_url, wait_monaco=False)

        hamburger = page.locator("#btn-toggle-main-sidebar")
        assert hamburger.is_visible(), "Hamburger icon sumiu"
        hamburger.click()
        page.wait_for_timeout(500)
        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "02_sidebar_toggle.png"))
        print("[PASS] Sidebar toggle funciona")
        b.close()


def test_lab_files_mount_button_visible(client_url, crolab_server):
    """(3) Confirma Botão de Acesso a Pastas Locais do Cliente."""
    with sync_playwright() as p:
        b = p.chromium.launch(headless=True)
        page = b.new_page()
        _login_and_goto_lab(page, client_url, wait_monaco=False)

        btn = page.locator("#btn-mount-local-drive")
        assert btn.is_visible(), "Botão Mount Drive deveria existir na Sidebar Lab"
        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "03_mount_button.png"))
        print("[PASS] Mount Drive button visível")
        b.close()


def test_lab_notebook_toolbar(client_url, crolab_server):
    """(4) Verificar ferramentas primárias do editor."""
    with sync_playwright() as p:
        b = p.chromium.launch(headless=True)
        page = b.new_page()
        _login_and_goto_lab(page, client_url, wait_monaco=False)

        count = page.locator("button.notebook-toolbar-btn").count()
        assert count >= 2, f"Esperava >= 2 toolbar buttons, encontrou {count}"
        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "04_toolbar.png"))
        print(f"[PASS] Toolbar tem {count} botões")
        b.close()


def test_lab_runtime_status(client_url, crolab_server):
    """(5) Confirma Badge de Infraestrutura Ativa."""
    with sync_playwright() as p:
        b = p.chromium.launch(headless=True)
        page = b.new_page()
        _login_and_goto_lab(page, client_url, wait_monaco=False)

        badge = page.locator(".lab-runtime-status")
        assert badge.is_visible(), "Badge de runtime não encontrado"
        badge_text = badge.inner_text()
        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "05_status_badge.png"))
        print(f"[PASS] Runtime badge: '{badge_text}'")
        b.close()


# ═══════════════════════════════════════════════════════════════
#  TESTES 6-10: Kernel Stateful (BROWSER COMPARTILHADO)
#  Monaco CDN carrega UMA vez. Sessão WS persiste entre testes.
# ═══════════════════════════════════════════════════════════════

@pytest.fixture(scope="module")
def kernel_page(client_url, crolab_server):
    """Sessão persistente de browser para testes do Kernel."""
    pw = sync_playwright().start()
    browser = pw.chromium.launch(headless=True)
    page = browser.new_page(viewport={"width": 1440, "height": 900})
    _login_and_goto_lab(page, client_url, wait_monaco=True)
    yield page
    browser.close()
    pw.stop()


def test_kernel_stateful_variables(kernel_page):
    """(6) Valida Memória Persistente do Daemon Python."""
    page = kernel_page

    _inject_and_run(page, "GLOBAL_STATE = 8899")
    out = _inject_and_run(page, "print(f'Recuperado: {GLOBAL_STATE}')")

    print(f"[DEBUG] Output kernel stateful: {out[:200]}")
    assert "8899" in out, f"Falha Stateful: Kernel não lembrou. Output: {out[:200]}"
    page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "06_stateful_var.png"))
    print("[PASS] Kernel Stateful: variável persistiu entre células")


def test_kernel_stateful_imports(kernel_page):
    """(7) Valida Import Persistence."""
    page = kernel_page

    _inject_and_run(page, "import math")
    out = _inject_and_run(page, "print(math.sqrt(16))")

    print(f"[DEBUG] Output import: {out[:200]}")
    assert "4.0" in out, f"Import não persistiu. Output: {out[:200]}"
    page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "07_stateful_import.png"))
    print("[PASS] Import math persistiu entre células")


def test_kernel_error_traceback(kernel_page):
    """(8) Exceções não crasham o Daemon e voltam como stderr."""
    page = kernel_page

    out = _inject_and_run(page, "print(variavel_que_nao_existe)")

    print(f"[DEBUG] Output error: {out[:200]}")
    assert "NameError" in out, f"Erro sem NameError. Output: {out[:200]}"
    page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "08_error_trace.png"))
    print("[PASS] NameError retornado corretamente pelo Kernel")


def test_kernel_streaming_output(kernel_page):
    """(9) Valida Loop assíncrono streaming para o Frontend."""
    page = kernel_page

    out = _inject_and_run(page, "for i in range(3): print(f'stream_{i}')")

    print(f"[DEBUG] Output streaming: {out[:200]}")
    assert "stream_0" in out and "stream_2" in out, f"Streaming falhou. Output: {out[:200]}"
    page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "09_streaming.png"))
    print("[PASS] Streaming de loop funciona (3 linhas)")


def test_kernel_empty_cell_no_execution(kernel_page):
    """(10) Célula Vazia gera toast sem invocar WS."""
    page = kernel_page

    page.evaluate("window.monacoEditor.setValue('  ');")
    page.click("#btn-run-cell")
    page.wait_for_timeout(800)

    toast = page.locator("#toast")
    toast_text = toast.inner_text() if toast.is_visible() else ""
    print(f"[DEBUG] Toast text: '{toast_text}'")
    assert "vazia" in toast_text.lower(), f"Toast deveria conter 'vazia'. Obteve: '{toast_text}'"
    page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "10_empty_cell.png"))
    print("[PASS] Célula vazia gera toast corretamente")
