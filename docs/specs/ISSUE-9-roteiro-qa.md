# ISSUE-9 — Roteiro de QA

Spec: [ISSUE-9-forma-pagamento-despesas-fixas.md](./ISSUE-9-forma-pagamento-despesas-fixas.md). Snapshot final será registrado no fechamento.

## QA1 — Revisão humana de contrato e integridade · cobre P4 (T1)
- Por que manual: o revisor independente disponível tem ferramentas de escrita não restritas; a integração Codex da skill exige revisão humana registrada ou ambiente que imponha somente leitura para esta task de risco alto.
- Pré-condições: checkout do snapshot de implementação indicado no fechamento e comparação com `88a3c5bd27f99e4417a3f80bee8bc4a5eb358cc8`; acesso às evidências dos commits de T1/T2.
- Passos:
  1. Revisar migration 6, preenchimento de todas as guilds/status como crédito, NOT NULL/CHECK e ausência de default permanente.
  2. Revisar POST obrigatório, PATCH omitido preservando e null/inválido rejeitado; autorização, escopo de guild e persistência dos cinco valores.
  3. Conferir os testes de migração e de round-trip PostgreSQL, incluindo leitura após POST e cada PATCH, e registrar SHA e resultado da revisão.
- Resultado esperado: contrato da revisão 2 atendido e nenhum achado bloqueante aberto no escopo revisado.
- Resultado: ⬜ pendente — responsável: revisor humano.
