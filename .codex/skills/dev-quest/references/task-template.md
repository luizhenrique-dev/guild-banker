# Template — arquivo de task

Arquivo: `<dir>/<ISSUE>-T<N>-<slug>.md`. Deve ser executável sem depender da memória da conversa.

```markdown
# <ISSUE> — T<N>: <título curto>

Spec: [<ISSUE>-<slug>.md](./<ISSUE>-<slug>.md)

- **Status:** ⬜ pendente
- **Depende de:** T<x>, T<y> | nenhuma

## O que fazer
1. <passo objetivo>
2. <testes a escrever/atualizar — nomes marcados como (planejado)>

## Arquivos
- [alterar] `path/verificado.ext` — <o quê>
- [criar] `path/proposto.ext` — <o quê>

## Critérios de aceite
- T<N>-AC1 <condição verificável> → cobre CA<n>
- T<N>-AC2 …

## Verificação
- `<comando confirmado>` em `<dir>` — pré-requisitos: … — esperado: <observável>
- Manual: … | não aplicável

## Contexto adicional
<somente o estritamente necessário: contratos, decisões D<n> relevantes, padrões a seguir>

## Evidências
<!-- preenchido pelo dev-mage -->
```

Regras: definição idêntica à do índice da spec; estados ⬜ 🔄 ✅ ⛔; testes planejados nunca descritos como existentes; não inclua comandos exemplificativos não confirmados.
