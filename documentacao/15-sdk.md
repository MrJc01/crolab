*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 14-modelo-de-negocio.md](14-modelo-de-negocio.md)
<hr>

# 15 — SDK para Desenvolvedores

## Visão

O binário `crolab` não é apenas uma CLI — é um **SDK embarcável**. Os pacotes internos podem ser importados por qualquer aplicação Go para construir integrações, automações ou até o site oficial.

## Pacotes Públicos

```
github.com/crolab/core/internal/cloud   → API REST + SQLite + Auth
github.com/crolab/core/internal/cli     → Config management
github.com/crolab/core/internal/node    → gRPC job execution
github.com/crolab/core/internal/lab     → Lab web IDE server
github.com/crolab/core/internal/tui     → Terminal UI (monitor)
```

## Exemplo: Embarcando o Crolab no seu app

```go
package main

import (
    "log"
    "github.com/crolab/core/internal/cloud"
)

func main() {
    // 1. Inicializa o banco
    cloud.InitDB("minha-app.db")
    cloud.SeedDefaultMachines()

    // 2. Usa as funções diretamente
    user, _ := cloud.DBCreateUser("dev@app.com", "senha123", "admin")
    log.Printf("User criado: %s (token: %s)", user.Email, user.Token)

    // 3. Cria plano programaticamente
    cloud.DBCreatePlan(cloud.DBPlan{
        ID: "gpu-pro", Name: "GPU Pro",
        VRAM: "24GB", PriceHr: 0.70,
    })

    // 4. Sobe a API na porta que quiser
    cloud.StartCloudServer(":9000", "./meu-frontend")
}
```

## Exemplo: Usando o BuildMux para integrar em servidor existente

```go
package main

import (
    "net/http"
    "github.com/crolab/core/internal/cloud"
)

func main() {
    cloud.InitDB("app.db")

    // Monta o handler Crolab
    crolabHandler := cloud.BuildMux("./web/client")

    // Integra no seu mux
    mux := http.NewServeMux()
    mux.Handle("/api/", http.StripPrefix("/api", crolabHandler))
    mux.Handle("/", http.FileServer(http.Dir("./site")))

    http.ListenAndServe(":8080", mux)
}
```

## Exemplo: Automação via API HTTP

```bash
# Criar plano via CLI
crolab admin plan create gpu-start "GPU Start" --vram 6GB --price 0.30

# Ou via cURL
curl -X POST http://localhost:8844/admin/plans \
  -H "Authorization: $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"id":"gpu-start","name":"GPU Start","vram":"6GB","price_hr":0.30}'
```

## Roadmap SDK

| Fase | Descrição |
|---|---|
| v0.2 | Pacotes internos usáveis via import (atual) |
| v0.3 | Mover para `pkg/` (API pública estável) |
| v0.4 | Client SDK (Go + Python + JS) para consumir a API |
| v0.5 | Plugin system (providers customizados) |
| v1.0 | Go module publicado: `go get github.com/crolab/sdk` |

## Funções DB disponíveis

### Users
- `DBCreateUser(email, password, role)` → cria com bcrypt
- `DBGetUserByEmail(email)` / `DBGetUserByToken(token)` / `DBGetUserByID(id)`
- `DBUpdateCredits(id, delta)` / `DBUpdateUserCredits(id, abs)` / `DBUpdateUserRole(id, role)`
- `DBListUsers()` / `DBUserCount()` / `DBDeleteUser(id)`

### Plans
- `DBCreatePlan(plan)` / `DBGetPlan(id)` / `DBUpdatePlan(plan)` / `DBDeletePlan(id)` / `DBListPlans()`

### Pool
- `DBAddPoolEntry(entry)` / `DBListPool(planID)` / `DBDeletePoolEntry(id)`

### Machines
- `DBCreateMachine(m)` / `DBListMachines()` / `DBUpdateMachine(m)` / `DBDeleteMachine(id)`

### User Machines
- `DBAddUserMachine(...)` / `DBListUserMachines(userID)` / `DBDeleteUserMachine(userID, machineID)`

### Jobs
- `DBCreateJob(job)` / `DBListJobs(userID)` / `DBUpdateJobStatus(id, status)`

### Transactions
- `DBLogTransaction(userID, amount, type, desc)` / `DBListTransactions(userID)`

### Subscriptions
- `DBSubscribe(userID, planID)` / `DBGetSubscription(userID)` / `DBUnsubscribe(userID)`

### Auth Helpers
- `HashPassword(pw)` / `CheckPassword(hash, pw)` / `GenerateToken()`

<hr>
*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 14-modelo-de-negocio.md](14-modelo-de-negocio.md)
