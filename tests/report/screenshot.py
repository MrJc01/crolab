from playwright.sync_api import sync_playwright
import time
import sys
import os

target_dir = sys.argv[1]

try:
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1440, "height": 900})
        page.on("console", lambda msg: print(f"Browser console: {msg.text}"))
        page.on("pageerror", lambda exc: print(f"Browser error: {exc}"))

        print("Capturing Admin Panel (SSO)...")
        page.goto("http://localhost:18844")
        time.sleep(2)
        if page.locator('#login-email').is_visible():
            page.fill('#login-email', 'admin@crolab.com')
            page.fill('#login-password', 'admin123')
            page.click('#btn-login')
            time.sleep(2)
        
        for tab in ["dashboard", "plans", "machines", "users", "logs"]:
            print(f"Capturing Admin {tab}...")
            if page.locator(f'#btn-tab-{tab}').is_visible():
                page.click(f'#btn-tab-{tab}')
                time.sleep(1)
                page.screenshot(path=os.path.join(target_dir, f"admin_{tab}.png"))

        print("Capturing Client Panel (SSO)...")
        page.goto("http://localhost:18855")
        time.sleep(2)
        
        if page.locator('#btn-show-auth').is_visible():
            print("Landing visible. Going to auth proxy...")
            page.click("#btn-show-auth")
            time.sleep(1)
            
        if page.locator('#auth-email').is_visible():
            print("SSO Failed for Client, doing manual login...")
            page.fill('#auth-email', 'admin@crolab.com')
            page.fill('#auth-password', 'admin123')
            page.click('#btn-auth-login')
            time.sleep(2)
        else:
            print("✅ Web Client utilizou SSO local_token magicamente!")
            
        for tab in ["home", "plans", "machines", "jobs", "billing"]:
            print(f"Capturing Client {tab}...")
            if page.locator(f'#btn-tab-{tab}').is_visible():
                page.click(f'#btn-tab-{tab}')
                time.sleep(1)
                page.screenshot(path=os.path.join(target_dir, f"client_{tab}.png"))

        browser.close()
        print("✅ Screenshots generated!")

except Exception as e:
    print(f"Failed: {e}")
    sys.exit(1)
