"""
02 — Dashboard Client (Home)
Verifica métricas, cards, e integridade visual do Dashboard após login.
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


def test_dashboard_metrics_visible(client_url, crolab_server):
    """Verifica que os 3 cards de métricas aparecem no dashboard."""
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1440, "height": 900})
        _login(page, client_url)

        assert page.locator("#h-credits").is_visible(), "Card de créditos ausente"
        assert page.locator("#h-plan").is_visible(), "Card de plano ausente"
        assert page.locator("#h-machines").is_visible(), "Card de máquinas ausente"

        credits_text = page.locator("#h-credits").inner_text()
        assert "$" in credits_text, f"Créditos sem símbolo $: {credits_text}"

        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "01_dashboard_metrics.png"))
        browser.close()


def test_dashboard_terminal_integration_card(client_url, crolab_server):
    """Verifica que o bloco Terminal Integration aparece com exemplo de comando."""
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1440, "height": 900})
        _login(page, client_url)

        quick_start = page.locator(".quick-start")
        assert quick_start.is_visible(), "Quick Start card não apareceu"

        code_block = page.locator(".code-block code")
        assert code_block.is_visible(), "Bloco de código CLI ausente"
        assert "crolab run" in code_block.inner_text(), "Exemplo CLI incorreto"

        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "02_terminal_integration.png"))
        browser.close()


def test_dashboard_user_info_topbar(client_url, crolab_server):
    """Verifica que email e créditos aparecem na topbar após login."""
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1440, "height": 900})
        _login(page, client_url)

        email = page.locator("#user-email").inner_text()
        assert "admin@crolab.com" in email, f"Email incorreto na topbar: {email}"

        credits = page.locator("#user-credits").inner_text()
        assert "$" in credits, f"Créditos na topbar sem $: {credits}"

        logout_btn = page.locator("#btn-logout")
        assert logout_btn.is_visible(), "Botão de logout ausente"

        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "03_topbar_user.png"))
        browser.close()
