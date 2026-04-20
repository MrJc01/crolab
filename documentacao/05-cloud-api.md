*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 04-arquitetura.md](04-arquitetura.md) &nbsp; | &nbsp; [Próximo: 06-lab.md ➡️](06-lab.md)
<hr>

# Cloud API — Auth, Billing e Machines

## Endpoints

Base URL: `http://localhost:8844`

### Autenticação

#### POST `/auth/register`
```json
// Request
{ "email": "user@email.com", "password": "senha123" }

// Response 201
{ "token": "abc123...", "message": "Conta criada. 10 créditos de boas-vindas." }

// Response 409
{ "error": "email já registrado" }
```

#### POST `/auth/login`
```json
// Request
{ "email": "user@email.com", "password": "senha123" }

// Response 200
{ "token": "abc123...", "credits": 10.0 }

// Response 401
{ "error": "credenciais inválidas" }
```

### Billing

#### GET `/billing/status`
Header: `Authorization: <token>`

```json
// Response 200
{ "email": "user@email.com", "credits": 60.0 }
```

#### POST `/billing/purchase`
Header: `Authorization: <token>`

```json
// Request
{ "amount": 50.0 }

// Response 200
{ "credits": 60.0, "message": "50.00 créditos adicionados" }
```

### Machines

#### GET `/machines`
```json
// Response 200
[
  { "id": "crom-a100-01", "name": "A100-Brazil-01", "gpu": "A100", "vram": "80GB", "price_hr": 1.50, "status": "available" },
  { "id": "crom-4090-01", "name": "RTX4090-EU-01", "gpu": "RTX 4090", "vram": "24GB", "price_hr": 0.60, "status": "available" },
  { "id": "crom-t4-01", "name": "T4-US-01", "gpu": "T4", "vram": "16GB", "price_hr": 0.25, "status": "available" }
]
```

#### POST `/machines/rent`
Header: `Authorization: <token>`

```json
// Request
{ "machine_id": "crom-t4-01" }

// Response 200
{
  "machine": { "id": "crom-t4-01", "name": "T4-US-01", "status": "rented" },
  "message": "Máquina T4-US-01 ativada. Use: crolab config add T4-US-01 10.0.1.4:4422",
  "credits": 9.75
}

// Response 402
{ "error": "créditos insuficientes (precisa: 1.50, tem: 0.25)" }
```

### O Orquestrador (Control Plane)

#### POST `/client/run`
Header: `Authorization: <token>`

Emite um "Ticket" criptográfico delegando o serviço de túnel pesado ao nó de borda de acordo com o pool selecionado.

```json
// Request (usando ID do plano ou da Máquina)
{ "plan_id": "start" }

// Response 201
{
  "address": "10.0.1.4:4422",
  "token": "abc123valido-node",
  "job_id": "a9f8b2c",
  "message": "Ticket gerado, prossiga a gRPC Dial."
}
```

## Configurações TLS (HTTPS) e Embedded UI

Na ausência do argumento `--web`, todo frontend visual de administração e clientes acoplados no `go:embed` são evocados pela base interna `/`. Você pode ocultar o tráfego REST passando os certificados SSL.

```bash
# Apenas API (Com TLS Ativado)
./crolab cloud-serve --tls-cert fullchain.pem --tls-key privkey.pem

# Inicialização Modo Provedor Dual-Port (Admin em :8844, Client em :8855)
./crolab provider --admin-port :8844 --client-port :8855
```

<hr>
*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 04-arquitetura.md](04-arquitetura.md) &nbsp; | &nbsp; [Próximo: 06-lab.md ➡️](06-lab.md)
