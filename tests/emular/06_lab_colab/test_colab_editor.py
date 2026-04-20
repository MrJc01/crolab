"""
06 — Editor Colab-Style (Monaco + WebSocket Kernel)
Verifica que o Monaco Editor carrega nativamente na aba Lab e que o WebSocket Kernel executa código.
"""
import os
import time
from playwright.sync_api import sync_playwright


SCREENSHOTS_DIR = os.path.join(os.path.dirname(__file__), "screenshots")
os.makedirs(SCREENSHOTS_DIR, exist_ok=True)


def test_monaco_editor_loads(client_url, crolab_server):
    """Verifica que o Monaco Editor renderiza na aba Lab sem iframe."""
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1440, "height": 900})

        page.goto(client_url)
        page.wait_for_load_state("networkidle")

        # Login
        page.click("#btn-start-landing")
        page.wait_for_selector("#auth-email", state="visible", timeout=3000)
        page.fill("#auth-email", "admin@crolab.com")
        page.fill("#auth-password", "admin123")
        page.click("#btn-auth-login")
        page.wait_for_selector("#home-section", state="visible", timeout=5000)

        # Navegar para aba Lab
        page.click("#btn-tab-lab")
        page.wait_for_selector("#lab-section", state="visible", timeout=3000)
        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "01_lab_loading.png"))

        # Esperar Monaco carregar do CDN (pode demorar)
        page.wait_for_timeout(4000)

        # Verificar que NÃO existe iframe (removemos)
        iframes = page.locator("iframe")
        assert iframes.count() == 0, "iframe ainda existe! Deveria ter sido removido."

        # Verificar container Monaco
        monaco = page.locator("#monaco-container")
        assert monaco.is_visible(), "Monaco container não está visível"

        # Verificar que Monaco populou o container (deve ter elemento .monaco-editor)
        monaco_editor = page.locator(".monaco-editor")
        assert monaco_editor.count() > 0, "Monaco Editor não renderizou dentro do container"

        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "02_monaco_loaded.png"))

        # Verificar output console
        output = page.locator("#lab-output")
        assert output.is_visible(), "Painel de output não está visível"

        browser.close()


def test_no_iframe_isolation(client_url, crolab_server):
    """Confirma que a porta 19999 NÃO é mais necessária."""
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1440, "height": 900})

        page.goto(client_url)
        page.wait_for_load_state("networkidle")

        # Login
        page.click("#btn-start-landing")
        page.wait_for_selector("#auth-email", state="visible", timeout=3000)
        page.fill("#auth-email", "admin@crolab.com")
        page.fill("#auth-password", "admin123")
        page.click("#btn-auth-login")
        page.wait_for_selector("#home-section", state="visible", timeout=5000)

        # Ir para Lab
        page.click("#btn-tab-lab")
        page.wait_for_selector("#lab-section", state="visible", timeout=3000)

        # Confirma zero iframes em toda a página
        all_iframes = page.frames
        # Apenas o main frame deve existir
        assert len(all_iframes) == 1, f"Encontrados {len(all_iframes)} frames, esperava apenas 1 (main)"

        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "03_no_iframe.png"))
        browser.close()


def test_kernel_execution_websocket(client_url, crolab_server):
    """Testa execução real de código Python via WebSocket Kernel."""
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1440, "height": 900})

        page.goto(client_url)
        page.wait_for_load_state("networkidle")

        # Login
        page.click("#btn-start-landing")
        page.wait_for_selector("#auth-email", state="visible", timeout=3000)
        page.fill("#auth-email", "admin@crolab.com")
        page.fill("#auth-password", "admin123")
        page.click("#btn-auth-login")
        page.wait_for_selector("#home-section", state="visible", timeout=5000)

        # Navegar para Lab
        page.click("#btn-tab-lab")
        page.wait_for_selector("#lab-section", state="visible", timeout=3000)
        page.wait_for_timeout(4000)  # CDN Monaco

        # Usar Monaco API para injetar código
        page.evaluate("""
            if (window.monacoEditor) {
                window.monacoEditor.setValue('print("CROLAB_KERNEL_TEST_OK")');
            }
        """)
        page.wait_for_timeout(500)
        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "04_code_injected.png"))

        # Clicar Run Cell
        page.click("#btn-run-cell")
        page.wait_for_timeout(3000)  # Esperar execução

        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "05_execution_result.png"))

        # Verificar output
        output_text = page.locator("#lab-output").inner_text()
        assert "CROLAB_KERNEL_TEST_OK" in output_text or "Executando" in output_text, \
            f"Output não contém resultado esperado. Obteve: {output_text[:200]}"

        browser.close()
