*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 12-testes-automatizados.md](12-testes-automatizados.md) &nbsp; | &nbsp; [Próximo: 14-modelo-de-negocio.md ➡️](14-modelo-de-negocio.md)
<hr>

# 13 — Provider Mode & Sincronização Cloud

A Crolab nasceu para democratizar o processamento de IA. Quando você hospeda um "Node Provider" na sua rede, você se torna uma Cloud privada (um mini GCP/AWS) ou um Hub P2P que aglutina e revende GPUs de outros cantos do planeta.

## Iniciação e o Segredo do 1º Login (First-Boot Wizard)
Você não precisa ler longos manuais para criar Senhas Mestras. A Engine faz isso por você.

1. Simplesmente inicie o provedor em **Background** pelo terminal:
   ```bash
   ./crolab provider start -d
   ```
2. O First-Boot Wizard detectará que a rede é virgem e injetará no Terminal as Suas Credenciais!
   ```text
   🎉 BEM-VINDO AO CROLAB PROVIDER!
   Usuário: root@crolab.local
   Senha:   cr0_a74f4b23
   ```
3. Acesse `http://localhost:8844` e faça o Login Master usando os dados forjados.

> [!TIP]
> Caso queira abortar a orquestração de fundo, digite: `./crolab provider stop`. Isso finalizará processos amigavelmente. Esse modelo `-d` se expande também para o `./crolab web start -d`.

### ⚡ Estratégia de Captação (Airdrop de Créditos)

Quer atrair pesquisadores para a sua Provider Cloud inicial? Você pode modificar a bonificação do Gateway com:
```bash
./crolab provider start --free-credits 50.0
# Ou exportando CROLAB_FREE_CREDITS="50.0"
```
A verificação anti-fraude via IP impossibilita o "farm de saldo" infinito pelas mesmas fontes de rede simultânea.

## Arbitragem e VastAI Sync

Fomos radicais: A API P2P de Cloud foi ativada.
Em vez de fakes e mockups, abra seu dashboard, localize no canto inferior o **☁️ Nuvem / Sync** (Provedores).

Clicar em **Sincronizar Mercado P2P** fará a Engine Crolab sugar via HTTP as Máquinas Fisicamente Existentes em infraP2Ps como a **Vast.AI**, pegará seus Custos Horários de Dólar e duplicará como uma Margem de Negócio dentro do seu banco de Dados (tornando elas disponíveis para os usuários Crolab `crolab client` alugarem agora).

<hr>
*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 12-testes-automatizados.md](12-testes-automatizados.md) &nbsp; | &nbsp; [Próximo: 14-modelo-de-negocio.md ➡️](14-modelo-de-negocio.md)
