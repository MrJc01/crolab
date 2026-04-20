"""
05 — Fila de Jobs (Execuções)
Verifica renderização da tabela de jobs e sincronização.
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


def test_jobs_page_renders(client_url, crolab_server):
    """Verifica que a página de jobs renderiza com tabela válida."""
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1440, "height": 900})
        _login(page, client_url)

        page.click("#btn-tab-jobs")
        page.wait_for_selector("#jobs-section", state="visible", timeout=3000)
        page.wait_for_timeout(1000)

        table = page.locator("#jobs-section .data-table")
        assert table.is_visible(), "Tabela de jobs não renderizou"

        thead = page.locator("#jobs-section thead th")
        assert thead.count() >= 5, f"Tabela deve ter >= 5 colunas, obteve {thead.count()}"

        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "01_jobs_table.png"))
        browser.close()


def test_jobs_refresh_button(client_url, crolab_server):
    """Verifica que o botão de sincronizar fila existe e é clicável."""
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1440, "height": 900})
        _login(page, client_url)

        page.click("#btn-tab-jobs")
        page.wait_for_selector("#jobs-section", state="visible", timeout=3000)

        refresh = page.locator("#btn-refresh-jobs")
        assert refresh.is_visible(), "Botão de refresh não encontrado"

        # Clicar deve não causar erro
        refresh.click()
        page.wait_for_timeout(500)

        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "02_jobs_refreshed.png"))
        browser.close()
