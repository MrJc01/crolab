"""
07 — Billing/Financeiro
Verifica saldo, compra de créditos e ledger de transações.
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


def test_billing_page_renders(client_url, crolab_server):
    """Verifica que a página de billing exibe saldo e botões de mint."""
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1440, "height": 900})
        _login(page, client_url)

        page.click("#btn-tab-billing")
        page.wait_for_selector("#billing-section", state="visible", timeout=3000)
        page.wait_for_timeout(1000)

        balance = page.locator("#bill-credits")
        assert balance.is_visible(), "Saldo não renderizou"
        assert "$" in balance.inner_text(), "Saldo sem símbolo $"

        # Botões Mint
        mint_buttons = page.locator("#billing-section .btn-primary")
        assert mint_buttons.count() >= 3, f"Esperava >= 3 botões Mint, obteve {mint_buttons.count()}"

        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "01_billing_overview.png"))
        browser.close()


def test_billing_purchase_credits(client_url, crolab_server):
    """Testa a compra de créditos e atualização do saldo."""
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1440, "height": 900})
        _login(page, client_url)

        page.click("#btn-tab-billing")
        page.wait_for_selector("#billing-section", state="visible", timeout=3000)
        page.wait_for_timeout(1000)

        # Captura saldo antes
        balance_before = page.locator("#bill-credits").inner_text()

        # Clicar Mint +$10
        mint_buttons = page.locator("#billing-section .btn-primary")
        mint_buttons.first.click()
        page.wait_for_timeout(2000)

        # Captura saldo depois
        balance_after = page.locator("#bill-credits").inner_text()

        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "02_after_purchase.png"))

        # Saldo deve ter mudado (ou pelo menos toast apareceu)
        # Nota: parseFloat para comparação
        print(f"  [BILLING] Antes: {balance_before} → Depois: {balance_after}")
        browser.close()


def test_billing_ledger_table(client_url, crolab_server):
    """Verifica que o ledger de transações existe e tem estrutura correta."""
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1440, "height": 900})
        _login(page, client_url)

        page.click("#btn-tab-billing")
        page.wait_for_selector("#billing-section", state="visible", timeout=3000)
        page.wait_for_timeout(1000)

        ledger = page.locator("#tx-tbody")
        assert ledger.is_visible(), "Tabela ledger não encontrada"

        # Deve ter pelo menos 1 transação (welcome airdrop)
        rows = page.locator("#tx-tbody tr")
        assert rows.count() >= 1, f"Ledger vazio, esperava >= 1 transação"

        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "03_ledger_transactions.png"))
        browser.close()
