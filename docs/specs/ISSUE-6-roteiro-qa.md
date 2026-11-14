# ISSUE-6 — Roteiro de QA

Spec: [ISSUE-6-tema-login-keycloak.md](./ISSUE-6-tema-login-keycloak.md).
Branch: `feat/ISSUE-6-tema-login-keycloak`. Snapshot: commit com `Task-ID: ISSUE-6/T1` e fechamento da ISSUE-6; registrar o SHA efetivamente avaliado abaixo.

Ambiente: desenvolvimento local, Keycloak 26.8.2 com o tema `guild-banker`.
Pré-requisitos: `make keycloak/up`, SPA disponível, usuário local de teste e senha configurada fora do Git. A suíte completa também requer API e banco preparados. Não gravar senha, código OIDC ou tokens em screenshots/relatórios.

## QA1 — Identidade visual e uso em mobile/desktop · CA1, CA2 (T1-AC1, T1-AC2), P5
- Por que manual: a spec exige avaliação humana da fidelidade ao design e clareza dos estados; testes automatizados verificam propriedades, navegação básica e ausência de overflow, mas não aprovam a experiência visual.
- Passos:
  1. Em sessão privada, abrir a SPA sem autenticação para chegar ao login.
  2. Conferir em 1440px e 375px com `web/prototype/DESIGN.md` e a composição de `web/prototype/screens/login.html` (formulário real substitui botões de demo/redirecionamento).
  3. Conferir marca, Inter, CTA azul pill, campos arredondados, painel escuro no desktop, formulário legível no mobile e ausência de rolagem horizontal.
  4. Navegar usando Tab/Shift+Tab; conferir ordem, foco visível e labels. Usar leitor de tela para conferir a associação da mensagem aos campos.
  5. Enviar senha inválida e conferir se a mensagem orienta a tentativa sem depender somente de cor. Não capturar o formulário com valores preenchidos.
  6. Abrir uma URL de autorização com `redirect_uri` inválido em ambiente de teste e conferir página de erro: marca, mensagem compreensível e somente ações nativas disponíveis, sem link para o redirect rejeitado.
- Resultado esperado: apresentação consistente com a referência, leitura confortável e erros compreensíveis/acessíveis nas duas larguras.
- Resultado: ⬜ pendente — analista, data e SHA a registrar.

## QA2 — Revisão humana do snapshot de risco alto · CA3, CA5, CA6 (T1-AC3, T2-AC2, T2-AC3)
- Por que manual: os subagentes revisores desta sessão possuem ferramentas de escrita; o harness não impõe a restrição de somente leitura exigida pelo pacote dev-flow para validar formalmente tasks de risco alto. A revisão independente por agente continua sendo evidência parcial.
- Passos:
  1. Registrar o SHA analisado e revisar o delta da ISSUE-6 contra `07228341b1f33a39f0f5aa144954a97ff3bd6454`.
  2. Conferir templates contra Keycloak 26.8.2: ações/URLs/scripts nativos preservados e overrides restritos a login/erros.
  3. Conferir a seleção pontual de `loginTheme`, ausência de overwrite/reset/importação dos demais recursos e operação sem escrita quando o tema já está correto.
  4. Conferir credenciais por ambiente, sessão temporária restrita e removida, timeout e mensagens sem segredos. Conferir a proteção dos diagnósticos Playwright.
  5. Examinar evidências dos commits de T1/T2 e do fechamento; confirmar que correspondem ao snapshot revisado e registrar eventual achado.
- Resultado esperado: revisão humana registrada com escopo e SHA, sem achados impeditivos abertos. Resultado positivo não substitui QA1.
- Resultado: ⬜ pendente — revisor, data e SHA a registrar.
