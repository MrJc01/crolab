"""
Crolab E2E Playwright Test Suite — Configuração Global
Sobe o Provider Server, autentica, e disponibiliza fixtures reutilizáveis.
"""
import pytest
import subprocess
import time
import os
import requests
import signal

ADMIN_PORT = 18844
CLIENT_PORT = 18855
DB_PATH = "/tmp/crolab_playwright_test.db"
BIN_PATH = os.path.join(os.path.dirname(__file__), "..", "..", "crolab")
ADMIN_EMAIL = "admin@crolab.com"
ADMIN_PASS = "admin123"
CLIENT_EMAIL = "dev@crolab.com"
CLIENT_PASS = "dev123"

_server_proc = None


def _wait_server(port, timeout=10):
    """Aguarda servidor HTTP ficar pronto."""
    for _ in range(timeout * 10):
        try:
            r = requests.get(f"http://localhost:{port}/metrics", timeout=0.3)
            if r.status_code == 200:
                return True
        except Exception:
            pass
        time.sleep(0.1)
    raise TimeoutError(f"Servidor na porta {port} não respondeu em {timeout}s")


@pytest.fixture(scope="session", autouse=True)
def crolab_server():
    """Sobe o Provider do Crolab uma vez para toda a sessão de testes."""
    global _server_proc

    # Limpa DB anterior
    if os.path.exists(DB_PATH):
        os.remove(DB_PATH)

    env = os.environ.copy()
    env["CROLAB_ADMIN_EMAIL"] = ADMIN_EMAIL
    env["CROLAB_ADMIN_PASS"] = ADMIN_PASS
    env["CROLAB_NO_PROMPT"] = "true"

    _server_proc = subprocess.Popen(
        [
            BIN_PATH, "provider", "start",
            "--admin-port", f":{ADMIN_PORT}",
            "--client-port", f":{CLIENT_PORT}",
            "--db", DB_PATH,
            "--no-prompt",
        ],
        env=env,
        cwd=os.path.join(os.path.dirname(__file__), "..", ".."),  # raiz do projeto
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
    )

    _wait_server(CLIENT_PORT)
    yield _server_proc

    _server_proc.send_signal(signal.SIGTERM)
    _server_proc.wait(timeout=5)


@pytest.fixture(scope="session")
def admin_token():
    """Autentica o admin e retorna o token."""
    resp = requests.post(
        f"http://localhost:{ADMIN_PORT}/auth/login",
        json={"email": ADMIN_EMAIL, "password": ADMIN_PASS},
    )
    assert resp.status_code == 200, f"Admin login falhou: {resp.text}"
    return resp.json()["token"]


@pytest.fixture(scope="session")
def client_token():
    """Registra um cliente e retorna o token."""
    resp = requests.post(
        f"http://localhost:{CLIENT_PORT}/auth/register",
        json={"email": CLIENT_EMAIL, "password": CLIENT_PASS},
        headers={"X-Forwarded-For": "192.168.50.1"},
    )
    assert resp.status_code == 201, f"Client register falhou: {resp.text}"
    return resp.json()["token"]


@pytest.fixture(scope="session")
def admin_url():
    return f"http://localhost:{ADMIN_PORT}"


@pytest.fixture(scope="session")
def client_url():
    return f"http://localhost:{CLIENT_PORT}"
