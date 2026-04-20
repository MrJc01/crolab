# Relatório E2E Playwright — Crolab Platform

> Gerado automaticamente pela suíte `tests/emular/run_all.sh`

| Campo | Valor |
|---|---|
| **Data** | 2026-04-20 16:05:06 |
| **Total de Testes** | 20 |
| **Aprovados** | ✅ 20 |
| **Falhas** | ❌ 0
0 |
| **Duração** | 284.37s |
| **Binário** | N/A |

---

## Módulos Testados

### 01 — Autenticação (Auth)
Verifica Landing Page, Login SSO, Registro de novo usuário.

![01_landing](../../tests/emular/01_auth/screenshots/01_landing.png)

![02_auth_form](../../tests/emular/01_auth/screenshots/02_auth_form.png)

![03_dashboard_post_login](../../tests/emular/01_auth/screenshots/03_dashboard_post_login.png)

![04_register_success](../../tests/emular/01_auth/screenshots/04_register_success.png)

### 02 — Dashboard (Home)
Verifica métricas de créditos, plano ativo, máquinas, e integração terminal.

![01_dashboard_metrics](../../tests/emular/02_dashboard/screenshots/01_dashboard_metrics.png)

![02_terminal_integration](../../tests/emular/02_dashboard/screenshots/02_terminal_integration.png)

![03_topbar_user](../../tests/emular/02_dashboard/screenshots/03_topbar_user.png)

### 03 — Planos e Assinaturas
Verifica catálogo de planos GPU, cards de preço, e fluxo de assinatura.

![01_plans_catalog](../../tests/emular/03_plans/screenshots/01_plans_catalog.png)

![02_subscription_info](../../tests/emular/03_plans/screenshots/02_subscription_info.png)

### 04 — Máquinas P2P (Nodos GPU)
Verifica grid de GPUs disponíveis, formulário Bridge RPC, e botões de aluguel.

![01_machines_grid](../../tests/emular/04_machines/screenshots/01_machines_grid.png)

![02_bridge_form](../../tests/emular/04_machines/screenshots/02_bridge_form.png)

![03_rent_buttons](../../tests/emular/04_machines/screenshots/03_rent_buttons.png)

### 05 — Fila de Jobs (Execuções)
Verifica tabela de jobs, colunas da fila, e botão de sincronização.

![01_jobs_table](../../tests/emular/05_jobs/screenshots/01_jobs_table.png)

![02_jobs_refreshed](../../tests/emular/05_jobs/screenshots/02_jobs_refreshed.png)

### 06 — Editor Colab-Style (Monaco + WebSocket Kernel)
Verifica que o Monaco Editor renderiza nativamente (sem iframe), que o WebSocket Kernel executa Python, e que os logs aparecem em tempo real.

![01_lab_loading](../../tests/emular/06_lab_colab/screenshots/01_lab_loading.png)

![02_monaco_loaded](../../tests/emular/06_lab_colab/screenshots/02_monaco_loaded.png)

![03_no_iframe](../../tests/emular/06_lab_colab/screenshots/03_no_iframe.png)

![04_code_injected](../../tests/emular/06_lab_colab/screenshots/04_code_injected.png)

![05_execution_result](../../tests/emular/06_lab_colab/screenshots/05_execution_result.png)

### 07 — Financeiro (Billing)
Verifica saldo de créditos, compra via botões Mint, e ledger de transações.

![01_billing_overview](../../tests/emular/07_billing/screenshots/01_billing_overview.png)

![02_after_purchase](../../tests/emular/07_billing/screenshots/02_after_purchase.png)

![03_ledger_transactions](../../tests/emular/07_billing/screenshots/03_ledger_transactions.png)

### 08 — Painel Admin
Verifica navegação completa entre Dashboard, Planos, Máquinas, Usuários e Logs.

![admin_dashboard](../../tests/emular/08_admin/screenshots/admin_dashboard.png)

![admin_logs](../../tests/emular/08_admin/screenshots/admin_logs.png)

![admin_machines](../../tests/emular/08_admin/screenshots/admin_machines.png)

![admin_plans](../../tests/emular/08_admin/screenshots/admin_plans.png)

![admin_users](../../tests/emular/08_admin/screenshots/admin_users.png)


---

## Log Completo da Execução

```
============================= test session starts ==============================
platform linux -- Python 3.12.3, pytest-9.0.3, pluggy-1.6.0 -- /tmp/crolab_screens/venv/bin/python
cachedir: .pytest_cache
rootdir: /home/j/Documentos/GitHub/crolab/tests/emular
collecting ... collected 20 items

01_auth/test_auth_flow.py::test_landing_page_renders   [browser] Failed to load resource: the server responded with a status of 401 (Unauthorized)
  [browser] Failed to load resource: the server responded with a status of 401 (Unauthorized)
  [DEBUG] Sections: [{'id': 'landing-section', 'hidden': False, 'display': 'block'}, {'id': 'auth-section', 'hidden': True, 'display': 'none'}, {'id': 'home-section', 'hidden': True, 'display': 'none'}, {'id': 'plans-section', 'hidden': True, 'display': 'none'}, {'id': 'machines-section', 'hidden': True, 'display': 'none'}, {'id': 'jobs-section', 'hidden': True, 'display': 'none'}, {'id': 'lab-section', 'hidden': True, 'display': 'none'}, {'id': 'billing-section', 'hidden': True, 'display': 'none'}, {'id': 'settings-section', 'hidden': True, 'display': 'none'}]
  [DEBUG] Page errors: []
PASSED
01_auth/test_auth_flow.py::test_login_flow PASSED
01_auth/test_auth_flow.py::test_register_new_user PASSED
02_dashboard/test_dashboard.py::test_dashboard_metrics_visible PASSED
02_dashboard/test_dashboard.py::test_dashboard_terminal_integration_card PASSED
02_dashboard/test_dashboard.py::test_dashboard_user_info_topbar PASSED
03_plans/test_plans.py::test_plans_page_renders PASSED
03_plans/test_plans.py::test_subscription_flow PASSED
04_machines/test_machines.py::test_machines_page_renders   [MACHINES] Cards encontrados: 0
PASSED
04_machines/test_machines.py::test_connect_form_toggle PASSED
04_machines/test_machines.py::test_gpu_card_has_rent_button   [MACHINES] Botões de aluguel encontrados: 0
PASSED
05_jobs/test_jobs.py::test_jobs_page_renders PASSED
05_jobs/test_jobs.py::test_jobs_refresh_button PASSED
06_lab_colab/test_colab_editor.py::test_monaco_editor_loads PASSED
06_lab_colab/test_colab_editor.py::test_no_iframe_isolation PASSED
06_lab_colab/test_colab_editor.py::test_kernel_execution_websocket PASSED
07_billing/test_billing.py::test_billing_page_renders PASSED
07_billing/test_billing.py::test_billing_purchase_credits   [BILLING] Antes: $0.00 → Depois: $10.00
PASSED
07_billing/test_billing.py::test_billing_ledger_table PASSED
08_admin/test_admin_panel.py::test_admin_dashboard_tabs PASSED

======================== 20 passed in 284.37s (0:04:44) ========================
```

---

## Inventário de Screenshots

| Arquivo | Módulo |
|---|---|
| `01_auth/screenshots/01_landing.png` | 01_auth |
| `01_auth/screenshots/02_auth_form.png` | 01_auth |
| `01_auth/screenshots/03_dashboard_post_login.png` | 01_auth |
| `01_auth/screenshots/04_register_success.png` | 01_auth |
| `02_dashboard/screenshots/01_dashboard_metrics.png` | 02_dashboard |
| `02_dashboard/screenshots/02_terminal_integration.png` | 02_dashboard |
| `02_dashboard/screenshots/03_topbar_user.png` | 02_dashboard |
| `03_plans/screenshots/01_plans_catalog.png` | 03_plans |
| `03_plans/screenshots/02_subscription_info.png` | 03_plans |
| `04_machines/screenshots/01_machines_grid.png` | 04_machines |
| `04_machines/screenshots/02_bridge_form.png` | 04_machines |
| `04_machines/screenshots/03_rent_buttons.png` | 04_machines |
| `05_jobs/screenshots/01_jobs_table.png` | 05_jobs |
| `05_jobs/screenshots/02_jobs_refreshed.png` | 05_jobs |
| `06_lab_colab/screenshots/01_lab_loading.png` | 06_lab_colab |
| `06_lab_colab/screenshots/02_monaco_loaded.png` | 06_lab_colab |
| `06_lab_colab/screenshots/03_no_iframe.png` | 06_lab_colab |
| `06_lab_colab/screenshots/04_code_injected.png` | 06_lab_colab |
| `06_lab_colab/screenshots/05_execution_result.png` | 06_lab_colab |
| `07_billing/screenshots/01_billing_overview.png` | 07_billing |
| `07_billing/screenshots/02_after_purchase.png` | 07_billing |
| `07_billing/screenshots/03_ledger_transactions.png` | 07_billing |
| `08_admin/screenshots/admin_dashboard.png` | 08_admin |
| `08_admin/screenshots/admin_logs.png` | 08_admin |
| `08_admin/screenshots/admin_machines.png` | 08_admin |
| `08_admin/screenshots/admin_plans.png` | 08_admin |
| `08_admin/screenshots/admin_users.png` | 08_admin |

> Relatório gerado em: 2026-04-20 16:05:06
