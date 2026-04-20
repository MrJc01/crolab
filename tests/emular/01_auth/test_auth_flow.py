"""
01 — Fluxo de Autenticação Completo (Register → Login → SSO)
Verifica renderização da landing page, formulário de login e dashboard pos-auth.
"""
import os
from playwright.sync_api import sync_playwright


SCREENSHOTS_DIR = os.path.join(os.path.dirname(__file__), "screenshots")
os.makedirs(SCREENSHOTS_DIR, exist_ok=True)


def test_landing_page_renders(client_url, crolab_server):
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1440, "height": 900})

        errors = []
        page.on("pageerror", lambda exc: errors.append(str(exc)))
        page.on("console", lambda msg: print(f"  [browser] {msg.text}"))

        # Limpa qualquer storage local para garantir estado limpo
        page.goto(client_url)
        page.evaluate("localStorage.clear()")
        page.reload()
        page.wait_for_load_state("networkidle")
        page.wait_for_timeout(3000)  # Espera JS checkAuth() resolver

        # Debug: dump HTML das sections
        debug = page.evaluate("""
            () => {
                const all = document.querySelectorAll('.page');
                return Array.from(all).map(el => ({
                    id: el.id,
                    hidden: el.classList.contains('hidden'),
                    display: getComputedStyle(el).display
                }));
            }
        """)
        print(f"  [DEBUG] Sections: {debug}")
        print(f"  [DEBUG] Page errors: {errors}")

        # Pelo menos uma seção principal deve estar visível
        visible = [s for s in debug if not s['hidden']]
        assert len(visible) > 0, f"Nenhuma seção visível. Debug: {debug}"

        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "01_landing.png"))
        browser.close()


def test_login_flow(client_url, crolab_server):
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1440, "height": 900})

        # Limpa storage para forçar landing
        page.goto(client_url)
        page.evaluate("localStorage.clear()")
        page.reload()
        page.wait_for_load_state("networkidle")

        # Ir para auth (clicar landing ou show-auth)
        if page.locator("#btn-start-landing").is_visible():
            page.click("#btn-start-landing")
        elif page.locator("#btn-show-auth").is_visible():
            page.click("#btn-show-auth")

        page.wait_for_selector("#auth-email", state="visible", timeout=5000)
        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "02_auth_form.png"))

        # Preencher login
        page.fill("#auth-email", "admin@crolab.com")
        page.fill("#auth-password", "admin123")
        page.click("#btn-auth-login")

        # Esperar dashboard carregar
        page.wait_for_selector("#home-section", state="visible", timeout=5000)
        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "03_dashboard_post_login.png"))

        # Verificar elementos do dashboard
        assert page.locator("#h-credits").is_visible(), "Créditos não apareceram"
        assert page.locator("#user-email").is_visible(), "Email do user não apareceu"

        browser.close()


def test_register_new_user(client_url, crolab_server):
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1440, "height": 900})

        # Limpa storage
        page.goto(client_url)
        page.evaluate("localStorage.clear()")
        page.reload()
        page.wait_for_load_state("networkidle")

        if page.locator("#btn-start-landing").is_visible():
            page.click("#btn-start-landing")
        elif page.locator("#btn-show-auth").is_visible():
            page.click("#btn-show-auth")

        page.wait_for_selector("#auth-email", state="visible", timeout=5000)

        # Switch para registro
        page.click("#btn-show-register")
        page.wait_for_selector("#reg-email", state="visible", timeout=2000)

        page.fill("#reg-email", "newuser@test.ai")
        page.fill("#reg-password", "testpass123")
        page.click("#btn-auth-register")

        page.wait_for_selector("#home-section", state="visible", timeout=5000)
        page.screenshot(path=os.path.join(SCREENSHOTS_DIR, "04_register_success.png"))

        browser.close()
