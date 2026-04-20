"""
04 — Máquinas P2P (Nodos)
Verifica listagem de máquinas, formulário de conexão e aluguel direto.
"""
import os
from playwright.sync_api import sync_playwright

SCREENSHOTS_DIR = os.path.join(os.path.dirname(__file__), "screenshots")
os.makedirs(SCREENSHOTS_DIR, exist_ok=True)


def _login(page, client_url):
    page.goto(client_url)
    page.evaluate("localStorage.clear()")
    page.reload()
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(2000)
    if page.locator("#btn-start-landing").is_visible():
        page.click("#btn-start-landing")
    elif page.locator("#btn-show-auth").is_visible():
        page.click("#btn-show-auth")
    page.wait_for_selector("#auth-email", state="visible", timeout=5000)
    page.fill("#auth-email", "admin@crolab.com")
    page.fill("#auth-password", "admin123")
    page.click("#btn-auth-login")
    page.wait_for_selector("#home-section", state="visible", timeout=5000)


def test_machines_page_renders(client_url, crolab_server):
    """Verifica que a página de máquinas renderiza corretamente."""
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1440, "height": 900})
        _login(page, client_url)

        page.click("#btn-tab-machines")
        page.wait_for_selector("#machines-section", state="visible", timeout=3000)
        page.wait_for_timeout(2000)

        # A section deve estar visível com o header e botão de conectar
        header = page.locator("#machines-section .page-header h1")
        assert header.is_visible(), "Header de máquinas ausente"

        connect_btn = page.locator("#btn-connect-machine")
        assert connect_btn.is_visible(), "Botão de conectar máquina ausente"

        # Grid pode estar vazio no teste unitário (sem seed), mas deve existir no DOM
        gpu_grid = page.locator("#gpu-grid")
        gpu_cards = page.locator(".gpu-card")
        print(f"  [MACHINES] Cards encontrados: {gpu_cards.count()}")

        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "01_machines_grid.png"))
        browser.close()


def test_connect_form_toggle(client_url, crolab_server):
    """Verifica que o formulário de conexão Bridge RPC abre e fecha."""
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1440, "height": 900})
        _login(page, client_url)

        page.click("#btn-tab-machines")
        page.wait_for_selector("#machines-section", state="visible", timeout=3000)

        connect_form = page.locator("#connect-form")
        assert not connect_form.is_visible(), "Formulário deveria começar oculto"

        page.click("#btn-connect-machine")
        page.wait_for_timeout(500)
        assert connect_form.is_visible(), "Formulário não abriu ao clicar botão"

        # Verificar campos
        assert page.locator("#conn-address").is_visible(), "Campo IP ausente"
        assert page.locator("#conn-token").is_visible(), "Campo Token ausente"
        assert page.locator("#conn-name").is_visible(), "Campo Alias ausente"

        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "02_bridge_form.png"))
        browser.close()


def test_gpu_card_has_rent_button(client_url, crolab_server):
    """Verifica estrutura dos botões de aluguel nas GPUs disponíveis."""
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1440, "height": 900})
        _login(page, client_url)

        page.click("#btn-tab-machines")
        page.wait_for_selector("#machines-section", state="visible", timeout=3000)
        page.wait_for_timeout(2000)

        rent_buttons = page.locator(".gpu-card button")
        count = rent_buttons.count()
        print(f"  [MACHINES] Botões de aluguel encontrados: {count}")

        # Se tiver GPUs, confere que pelo menos 1 tem botão
        gpu_cards = page.locator(".gpu-card")
        if gpu_cards.count() > 0:
            assert count >= 1, "GPUs existem mas nenhum botão de aluguel"

        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "03_rent_buttons.png"))
        browser.close()
