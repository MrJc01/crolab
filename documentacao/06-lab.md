*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 05-cloud-api.md](05-cloud-api.md) &nbsp; | &nbsp; [Próximo: 07-monitor.md ➡️](07-monitor.md)
<hr>

# Crolab Lab — Editor Web

## O que é

O Lab é um notebook web estilo Google Colab que roda localmente. Um único comando abre o editor no navegador com:
- Explorador de arquivos
- Editor de código com tabs
- Terminal com output em tempo real via WebSocket

## Como usar

```bash
# Abre a pasta atual
./crolab lab .

# Abre uma pasta específica
./crolab lab /home/usuario/projeto

# Porta customizada
./crolab lab . --port :9000
```

O navegador abre automaticamente.

## Funcionalidades

### Explorador de Arquivos (sidebar esquerda)
- Navega subdiretórios
- Ícones por tipo de arquivo (🐍 .py, 📜 .js, 🔷 .go, etc)
- Esconde dotfiles
- Botão 📁 para trocar pasta raiz

### Editor (centro)
- Tabs para múltiplos arquivos
- JetBrains Mono (monospace)
- Tab insere 4 espaços
- Indicador de "● modificado" / "✓ salvo"

### Terminal (inferior)
- Output em tempo real via WebSocket
- Input para comandos livres
- Cores: stdout (branco), stderr (vermelho), exit code (verde/vermelho)

### Execução Automática

O botão ▶ detecta a linguagem pelo arquivo e executa:

| Extensão | Comando |
|---|---|
| `.py` | `python3 arquivo.py` |
| `.js` | `node arquivo.js` |
| `.go` | `go run arquivo.go` |
| `.sh` | `bash arquivo.sh` |
| `.ts` | `npx ts-node arquivo.ts` |
| outro | `cat arquivo` |

## Atalhos

| Atalho | Ação |
|---|---|
| `Ctrl+S` | Salvar arquivo |
| `Ctrl+Enter` | Salvar e executar |
| `Tab` (no editor) | Inserir 4 espaços |

## API Interna

O Lab expõe uma API REST local:

```
GET  /api/files?dir=.       → Lista arquivos
GET  /api/file?path=main.py → Lê conteúdo
POST /api/save               → Salva arquivo
GET  /api/dir                → Retorna pasta atual
POST /api/setdir             → Troca pasta raiz
WS   /api/exec               → WebSocket para execução
```

<hr>
*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 05-cloud-api.md](05-cloud-api.md) &nbsp; | &nbsp; [Próximo: 07-monitor.md ➡️](07-monitor.md)
