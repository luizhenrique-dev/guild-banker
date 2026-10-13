# ISSUE-1 — T6: Fixture de access token Keycloak real

Spec: [ISSUE-1-autenticacao-jwt-rs256.md](./ISSUE-1-autenticacao-jwt-rs256.md)

- **Status:** ✅ concluida
- **Depende de:** T2, T4, T5

## O que fazer

1. Preparar realm guild-banker em Keycloak LOCAL ISOLADO, com identidade fictícia, client de emissão de teste, audience guild-banker-api nos access tokens e realm roles member/admin. Usar versão compatível com Keycloak 26.7.5 declarada no compose. Não alterar realm/volume/usuários compartilhados; não executar migration da API. D4 autoriza fixture isolada, não configuração do ambiente pessoal.
2. Keycloak precisa emitir o token: capturar access token original, JWKS público e metadados de issuer/audience/kid/versão/instante. JWKS não fornece chave privada e não permite assinar um novo token do realm. Nunca exportar/commitar chave privada real, credencial pessoal ou JWT de usuário real. Arquivos contêm somente material fictício de teste e são identificados assim.
3. Documentar no README preparo/captura reproduzível com comandos efetivamente validados no ambiente isolado, requisitos e configuração do mapper de audience. Capturar token+JWKS coerentes; não editar payload/sub/exp de JWT já assinado. Não exigir automatização de Docker/Keycloak para toda execução da suíte.
4. Criar TestVerifierKeycloakFixture e TestAuthKeycloakFixture (planejados) com tag integration, ambos em infra/keycloak/verifier_keycloak_integration_test.go. Reutilizar validação de produção e ponto privado de construção do verifier criado em T2; esse pacote pode importar middleware/auth.Service para o replay HTTP sem ciclo. Fornecer apenas chave pública local e relógio dentro do intervalo válido em teste. Preservar issuer original; não usar SkipIssuerCheck/SkipClientIDCheck/SkipExpiryCheck/InsecureSkipSignatureCheck. Relógio após exp comprova rejeição. User lookup da suíte usa registro fictício ativo ligado ao sub da fixture, sem banco real.
5. Replay pode construir verifier com key set público capturado ou transportar apenas URLs de teste conhecidas para httptest, conforme ponto de injeção já utilizado em T2; nenhuma opção pública de bypass é introduzida. Discovery/rotação continuam cobertos pelos testes sintéticos de T2/T4. Provar assinatura/claims e passagem HTTP pelo mesmo middleware/serviço com 200 válido e 401 expirado.
6. Exigir todos os testes de fixture no alvo api/test/integration, sem skip condicional por ausência de Keycloak/arquivo/token. Captura inicial depende da infraestrutura isolada; execução posterior usa apenas os arquivos versionados. Não criar pipeline CI; CA9 remoto foi transferido por D5.

## Arquivos

- [criar] `api/internal/infra/keycloak/testdata/keycloak/access-token.jwt` — JWT original de teste fictício.
- [criar] `api/internal/infra/keycloak/testdata/keycloak/jwks.json` — somente chaves públicas.
- [criar] `api/internal/infra/keycloak/testdata/keycloak/metadata.json` — versão/issuer/audience/kid/instante reproduzível, sem credenciais.
- [criar] `api/internal/infra/keycloak/testdata/keycloak/README.md` — procedimento validado e isolamento.
- [criar] `api/internal/infra/keycloak/verifier_keycloak_integration_test.go` — compatibilidade do adapter e replay HTTP da fixture real, com relógio privado do mesmo pacote.

## Critérios de aceite

- T6-AC1 Fixture contém JWT emitido pelo Keycloak e JWKS público correspondente, com identidade fictícia e origem documentada; nenhum segredo pessoal/chave privada real → CA9 local, CA13.
- T6-AC2 Token real recebe 200 no handler de prova e principal/claims corretos no contexto em instante fixo válido → CA2, CA10.
- T6-AC3 Mesma fixture recebe 401 depois de exp sem modificar JWT nem desabilitar validação → CA4.
- T6-AC4 Alvo integration executa a fixture real sem Keycloak/PostgreSQL ativos, sem skips; cobertura local CA1–CA8 continua passando → CA9 local.

## Verificação

- Na raiz: `make api/vet`, `make api/lint`, `make api/test`, `make api/test/integration`, `make api/build` — Go compatível/race/dependências; alvo integration criado em T2. Esperado: fixture real válida/expirada e suíte local completa passam sem serviços externos.
- Manual inicial: captura em Keycloak isolado, conferência de correspondência kid/chave pública, RS256, issuer e aud, identidade fictícia e ausência de chave privada/credenciais nos arquivos. Procedimento será documentado após execução; comandos de captura ainda não são fatos verificados no planejamento.
- Gate CI: não aplicável nesta issue, conforme D5; não alegar atendimento integral ao API-CA9 original.

## Contexto adicional

D4 aprovada: fixture real capturada, relógio fixo apenas em teste, suíte independente de Keycloak ativo. D6: user local previamente cadastrado; aqui o lookup é stubado com usuário fictício ligado ao sub. O issuer local de referência é http://localhost:8080/realms/guild-banker, lido da fixture/config, não hardcoded na implementação. Esta task entrega compatibilidade externa, não uma task genérica de testes finais. Sem export/import destrutivo ou infra/clean no ambiente compartilhado.

## Evidências

T5 concluída no commit `36ee656`; correção final dos órfãos de T4 no commit `78cfc48`.

| Critério | Evidência observada | Resultado |
|---|---|---|
| T6-AC1 | Keycloak26.7.5 standalone/JDK21 em /tmp; realm importado em dev-mem; resposta original do token endpoint com RS256, aud guild-banker-api e usuário fictício fixture-member; JWKS público correspondente e README do procedimento | captura concluída |
| T6-AC2 | TestVerifierKeycloakFixture e TestAuthKeycloakFixture/valid: claims, sub consultado, Actor local42/e-mail fictício, HTTP200 em valid_at | PASS com race |
| T6-AC3 | Mesmos testes após expires_at+1: ReasonTokenExpired/HTTP401 e nenhum lookup; JWT original sem alteração | PASS com race |
| T6-AC4 | make api/test/integration completo após shutdown confirmado da instância isolada; arquivos obrigatórios, nenhum skip e nenhuma conexão PostgreSQL | PASS com race |

O ambiente não tinha acesso ao Docker; foi usada distribuição standalone26.7.5 com banco em memória, bind127.0.0.1:18880 e hostname http://localhost:8080. JRE extraído em /tmp precisou de truststore público local; o sandbox autorizou sockets. A primeira emissão sem scope basic omitiu sub e foi recusada na captura; corrigido somente o client isolado, sem alterar validação da API. Metadados registram captura UTC e intervalo original. Nenhum realm/volume pessoal, migration ou usuário real foi usado. Processo isolado encerrado com Ctrl-C e mensagem Keycloak stopped antes da suíte completa.

RED observado: arquivos de fixture ainda ausentes. Após captura real, verifier e replay HTTP válido/expirado passaram. Revisão independente confirmou metadados, public-only e assinatura RS256 com RSA2048 via cálculo independente; nenhuma opção Skip na implementação ou nos testes.

Gates finais observados na raiz: `make api/vet api/lint api/test api/build` PASS; `make api/test/integration` PASS após correção T4 e shutdown; golangci-lint global com integration PASS (0 issues); `make web/types web/build` PASS (Vite apenas aviso de bundle >500kB). Nenhuma falha detectada nas verificações finais executadas. Go1.26.2, GCC15, GNU Make4.4.1 e Node24.6.0 foram fornecidos por PATH/CC/caches em /tmp sem atualizar versões do projeto. CI não executado/criado conforme D5.
