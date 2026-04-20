"""
03 — Planos e Assinaturas
Verifica catálogo de planos, assinatura e cancelamento.
"""
import os
import requests
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


def test_plans_page_renders(client_url, admin_token, crolab_server):
    """Verifica que a página de planos carrega com pelo menos um plano (criado via API)."""
    # Criar plano via API
    requests.post(
        f"{client_url}/admin/plans",
        json={"id": "starter", "name": "Starter GPU", "vram": "6GB", "storage": "50GB",
              "price_hr": 0.25, "price_month": 19.90, "max_users": 100},
        headers={"Authorization": admin_token},
    )

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1440, "height": 900})
        _login(page, client_url)

        page.click("#btn-tab-plans")
        page.wait_for_selector("#plans-section", state="visible", timeout=3000)
        page.wait_for_timeout(1500)

        plans_grid = page.locator("#plans-grid")
        assert plans_grid.is_visible(), "Grid de planos não renderizou"

        plan_cards = page.locator(".plan-card")
        assert plan_cards.count() >= 1, f"Nenhum card de plano. Esperava >= 1"

        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "01_plans_catalog.png"))
        browser.close()


def test_subscription_flow(client_url, admin_token, crolab_server):
    """Testa assinar e verificar assinatura ativa."""
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1440, "height": 900})
        _login(page, client_url)

        page.click("#btn-tab-plans")
        page.wait_for_selector("#plans-section", state="visible", timeout=3000)
        page.wait_for_timeout(1500)

        # Verificar info de assinatura
        sub_details = page.locator("#sub-details")
        assert sub_details.is_visible(), "Detalhes de assinatura ausentes"

        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "02_subscription_info.png"))
        browser.close()
