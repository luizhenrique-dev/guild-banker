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


## QA2 — Conferência visual final · cobre CA3, CA5 (T2-AC1, T2-AC3)
- Por que manual: a spec exige avaliação visual pelo analista/usuário; testes e capturas do agente complementam, sem substituir essa aprovação.
- Pré-condições: SPA e API do snapshot final, migration 6 aplicada em ambiente de QA; membro de uma guild com despesas nas cinco formas. Usar dados de teste.
- Passos:
  1. Abrir Despesas fixas a 375 e 1280 px, nos temas claro e escuro.
  2. Criar/editar escolhendo cada forma, salvar e reabrir; conferir a coluna após categoria e os cinco badges com ícones, incluindo Vale/Ticket.
  3. Navegar por teclado no seletor e filtro; conferir foco, legibilidade dos novos controles, rolagem local da tabela e ausência de controles cortados.
  4. Combinar forma de pagamento, categoria e responsável; conferir total só de ativas, vazio, idioma inglês e limpeza do filtro na troca de guild.
- Resultado esperado: cada método é identificável por texto e ícone, novos controles legíveis nos dois temas e fluxo utilizável nas duas larguras. Testes automatizados dos quatro cenários responsivos passaram, incluindo cor do texto no tema escuro. Há contraste insuficiente preexistente em campos antigos (categoria/responsável/nome/valor no tema escuro); não atribuir isso à nova coluna/filtro nem ampliar o escopo sem decisão.
- Resultado: ⬜ pendente — responsável: analista/usuário.

## QA3 — Reexecutar integração de login Keycloak · cobre CA6 (gate final)
- Por que pendente: a suíte E2E completa executou 51 cenários; 50 passaram e o callback real falhou ao preencher senha. E2E_PASSWORD/KC_TEST_USER_PASSWORD não estavam definidas no ambiente. Código e teste Keycloak não foram alterados pela issue.
- Pré-condições: Keycloak local configurado e credenciais do usuário de teste fornecidas ao processo por mecanismo seguro, sem gravá-las no repositório ou no relatório.
- Passos:
  1. Configurar a variável de senha de teste e E2E_USERNAME se necessário, para o usuário do ambiente.
  2. Na raiz, executar `npm --prefix web/spa run test:e2e -- e2e/keycloak-theme.spec.ts -g 'valid credentials'`.
- Resultado esperado: login retorna callback com authorization code e troca de tokens aceita, conforme as assertivas existentes. Registrar snapshot e resultado; eventual falha nova exige diagnóstico antes de aprovar o gate.
- Resultado: ⬜ pendente — responsável: desenvolvedor/analista com credenciais de teste.
