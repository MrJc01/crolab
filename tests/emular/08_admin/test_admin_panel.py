"""
08 — Painel Admin Completo
Verifica que o admin consegue navegar entre Dashboard, Planos, Máquinas, Users e Logs.
"""
import os
from playwright.sync_api import sync_playwright


SCREENSHOTS_DIR = os.path.join(os.path.dirname(__file__), "screenshots")
os.makedirs(SCREENSHOTS_DIR, exist_ok=True)


def test_admin_dashboard_tabs(admin_url, crolab_server):
    """Testa navegação completa por todas as abas do admin."""
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1440, "height": 900})

        page.goto(admin_url)
        page.wait_for_load_state("networkidle")

        # Login admin
        if page.locator("#login-email").is_visible():
            page.fill("#login-email", "admin@crolab.com")
            page.fill("#login-password", "admin123")
            page.click("#btn-login")
            page.wait_for_timeout(2000)

        tabs = ["dashboard", "plans", "machines", "users", "logs"]
        for tab in tabs:
            btn = page.locator(f"#btn-tab-{tab}")
            if btn.is_visible():
                btn.click()
                page.wait_for_timeout(1000)
                page.screenshot(path=os.path.join(SCREENSHOTS_DIR, f"admin_{tab}.png"))

        browser.close()
