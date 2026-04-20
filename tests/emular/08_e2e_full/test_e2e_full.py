import pytest
import subprocess
from playwright.sync_api import Page, expect
import time
import os

BIN_PATH = "./crolab"

@pytest.fixture(scope="module", autouse=True)
def start_server():
    server = subprocess.Popen([BIN_PATH, "provider", "start", "--admin-port", ":61001", "--client-port", ":61002", "--db", "/tmp/crolab_e2e.db", "--no-prompt"])
    time.sleep(2)
    yield
    server.terminate()
    try:
        os.remove("/tmp/crolab_e2e.db")
    except:
        pass

def test_full_scenario_e2e(page: Page):
    # 1. Register & Auth
    page.goto("http://localhost:61002/")
    if page.locator("#btn-login-toggle").is_visible():
        page.locator("#btn-login-toggle").click()
    page.fill("#login-email", "e2e_test@crolab.com")
    page.click("#btn-login") # Auth creates user dynamically in Crolab if not exists
    time.sleep(1)

    # 2. Billing & Subscription Dashboard check
    expect(page.locator("#user-credits")).to_contain_text("10.00") # Airdrop credits

    # 3. Lab / Run Execution Check
    page.goto("http://localhost:61002/#lab")
    time.sleep(1)
    expect(page.locator("#lab-filename-input")).to_be_visible()

    # Create test code
    if page.locator("#btn-add-code").is_visible():
         page.click("#btn-add-code")
    
    # 4. Verifies Runtime Drops (Python, Node)
    expect(page.locator("#runtime-selector")).to_be_visible()
