*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 09-testes.md](09-testes.md) &nbsp; | &nbsp; [Próximo: 11-admin-panel.md ➡️](11-admin-panel.md)
<hr>

# Deploy em Produção

## Cross-Compile

```bash
make build-all
```

Gera binários em `dist/`:
- `crolab-linux-amd64` (23MB)
- `crolab-linux-arm64` (21MB)
- `crolab-darwin-amd64` (23MB)
- `crolab-darwin-arm64` (23MB)

*Nota:* Graças à diretiva `go:embed`, todos os binários descritos acima **carregam os sistemas de interface inteiros** (Web Client, Admin Panel e Lab Viewer) unificados no executável. O deploy de Frontends está obsoleto, o binário contém tudo.

## Deploy em VPS

### 1. Instalar automaticamente

Na VPS remota:
```bash
curl -sSL https://crolab.crom.run/install | bash
```

### 2. Instalar manualmente

```bash
# Copiar binário
scp dist/crolab-linux-amd64 root@vps:/usr/local/bin/crolab
ssh root@vps "chmod +x /usr/local/bin/crolab"

# Iniciar como node
ssh root@vps "crolab serve --port :4422 --token $(uuidgen) --slots 4 &"
```

### 3. Registrar no seu PC local

```bash
crolab config add minha-vps IP_DA_VPS:4422 TOKEN_GERADO --provider vps --priority 2
```

## Hijack de Instâncias Cloud

O script `scripts/install.sh` automatiza:
1. Detecta OS e arquitetura
2. Baixa binário
3. Instala em `/usr/local/bin/`
4. Gera token
5. Inicia daemon
6. Printa comando de conexão

## Systemd (Produção)

Para manter o daemon rodando:

```ini
# /etc/systemd/system/crolab.service
[Unit]
Description=Crolab Node
After=network.target docker.service

[Service]
Type=simple
ExecStart=/usr/local/bin/crolab serve --port :4422 --token YOUR_TOKEN --slots 4
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable crolab
sudo systemctl start crolab
```

## Métricas

O node expõe métricas nativas do Prometheus em `http://localhost:9090/metrics` (para Dataplane) e `/metrics` (na API REST Control Plane):

```prometheus
# HELP crolab_users_total Total de usuários registrados na plataforma.
# TYPE crolab_users_total gauge
crolab_users_total 12

# HELP crolab_machines_online_total Total de máquinas disponíveis ou rodando jobs.
# TYPE crolab_machines_online_total gauge
crolab_machines_online_total 8
```

Health check: `GET http://localhost:9090/health`
```json
{ "status": "ok" }
```

<hr>
*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 09-testes.md](09-testes.md) &nbsp; | &nbsp; [Próximo: 11-admin-panel.md ➡️](11-admin-panel.md)
