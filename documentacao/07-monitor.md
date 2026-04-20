*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 06-lab.md](06-lab.md) &nbsp; | &nbsp; [Próximo: 08-seguranca.md ➡️](08-seguranca.md)
<hr>

# Monitor TUI — Dashboard no Terminal

## O que é

O `crolab monitor` é um dashboard interativo que roda direto no terminal usando BubbleTea + Lipgloss.

## Como usar

```bash
./crolab monitor
```

## Interface

```
  CROLAB MONITOR

  ★ meu-gpu          local      1     192.168.1.10:4422    ● 12ms
    vast-backup       vastai     2     45.67.89.10:4422     ○ offline
    runpod-01         runpod     3     123.45.67.89:4422    ● 45ms

  ✓ Default → meu-gpu

  Logs
  ──────────────────────────────────────────────────
  │ Monitor iniciado.
  │ Default alterado para meu-gpu
  │ Removido: old-node

  Servers: 3  │  ↑↓ navegar  Enter=default  D=remover  A=adicionar  R=refresh  Q=sair
```

## Teclas

| Tecla | Ação |
|---|---|
| ↑ / ↓ | Navegar entre servidores |
| Enter | Definir selecionado como default (★) |
| D / Delete | Remover servidor selecionado |
| A | Abrir formulário para adicionar novo servidor |
| R | Refresh da lista + ping de todos |
| Q / Ctrl+C | Sair |

## Formulário de Adição (tecla A)

```
  Adicionar Servidor

  Nome:
  [meu-novo-gpu                 ]

  Endereço:
  [10.0.0.5:4422                ]

  Token:
  [                             ]

  Provider:
  [vastai                       ]

  Tab=próximo campo  Enter=salvar  Esc=cancelar
```

## Ping de Status

- Auto-refresh a cada 10 segundos
- Tenta gRPC dial com timeout de 2s
- `● 12ms` = online com latência
- `○ offline` = não respondeu

## Painel de Logs

Últimas 6 ações são mostradas:
- Alterações de default
- Remoções
- Adições
- Refreshes

<hr>
*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 06-lab.md](06-lab.md) &nbsp; | &nbsp; [Próximo: 08-seguranca.md ➡️](08-seguranca.md)
