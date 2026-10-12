# Contexto — escopo git e projeto

## Estabelecer o snapshot (Fase 0)

```bash
git rev-parse --show-toplevel
git status --short
git branch --show-current && git rev-parse HEAD
git show-ref --verify refs/heads/develop
git merge-base develop HEAD
git diff --name-status --find-renames develop...HEAD
git diff --find-renames --no-ext-diff --no-textconv develop...HEAD
```

Registre SHA da base, do ancestral comum e de HEAD. Se `develop` não existir ou a base for ambígua: **pergunte**. Não prossiga com outra base em silêncio.

**Alterações não commitadas:** escopo padrão são os commits da branch. Se houver working tree sujo e o pedido não esclarecer, confirme o alvo. Nunca misture commits da branch com alterações de working tree. Checks rodados em working tree diferente não validam o snapshot commitado.

## Ler contexto do projeto

Leia antes de julgar qualquer achado:

- Instruções: `AGENTS.md`, `CLAUDE.md`, `CONTRIBUTING`, READMEs de diretório
- Build/CI: scripts de build, configurações de CI/CD
- Manifests e lockfiles: `package.json`, `go.mod`, `requirements.txt` etc.
- ADRs/DAS: decisões arquiteturais registradas
- Testes relacionados ao delta
- Descrição/critérios de aceite da feature quando fornecidos

**Guidelines e convenções documentadas são a régua.** Divergência delas é achado; preferência pessoal não é.

## Classificação de arquivos no diff

| Tipo | Abordagem |
|---|---|
| Implementação / testes | Análise completa por passes A–F |
| Config / infra / pipeline | Análise proporcional: integridade e aderência ao padrão |
| Contratos / schemas / migrations | Análise proporcional: compatibilidade, reversibilidade |
| Dependências / lockfiles | Passe H obrigatório |
| Gerados / snapshots / vendor | Análise proporcional (versões, intencionalidade) — sem exclusão automática |

Arquivos mecânicos recebem análise proporcional, não exclusão.
