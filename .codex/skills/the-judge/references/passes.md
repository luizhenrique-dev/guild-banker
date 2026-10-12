# Passes de revisão (A–H)

Aplique os passes relevantes ao delta. Não produza comentário para preencher passe vazio. **G e H são obrigatórios.**

---

## A — Funcionamento e regras de negócio

Invariantes preservados, nulos/vazios/limites tratados, precisão (dinheiro, datas, fuso), erros parciais, idempotência, concorrência/transações, compatibilidade de contratos.

---

## B — Segurança

Authn/authz no recurso correto, isolamento de tenant, validação em fronteiras de entrada, injeções/path traversal, segredos em código/logs/respostas. **Demonstre** entrada → caminho → impacto. Ferramenta perigosa em abstrato não é exploração.

---

## C — Confiabilidade e operação

Timeouts, retries com backoff e idempotência, liberação de recursos, filas/reprocessamento, logs sem dados sensíveis, métricas/tracing no fluxo alterado, deploy gradual, migrations com rollback/roll-forward.

---

## D — Performance e custo

N+1, paginação/lotes, I/O bloqueante, memória, chamadas remotas, cache (invalidação/isolamento). **Risco demonstrável ≠ hipótese de carga.** Não invente percentuais.

---

## E — Design e convenções

Coesão, acoplamento, direção de dependências, consistência com o codebase e com ADRs, duplicação semântica, abstração especulativa, nomes. SOLID/DRY/KISS/YAGNI são heurísticas, não exigências.

---

## F — Testes

Caminhos alterados cobertos, assertivas observam o resultado certo, mocks que escondem o comportamento, testes frágeis, removidos, enfraquecidos ou ignorados. Sugira teste **só com cenário + resultado esperado**. Ausência de teste em mudança crítica pode ser ALTO; explique qual garantia falta.

---

## G — Desvio de escopo (obrigatório)

Liste **toda** alteração sem relação com o objetivo da feature: refactors oportunistas, renomeações, formatação em massa, mudanças em config/CI/infra, arquivos de outros domínios.

Para cada desvio:
- Aumenta risco ou superfície de regressão → 🔴 ALTO
- Não aumenta risco mas dificulta review/rollback → 🟡 MÉDIO (peça PR separado)
- Trivial e isolada → 🟢 BAIXO

Sem descrição do objetivo da feature: infira do nome da branch/commits e registre a inferência como ❓.

---

## H — Dependências externas (obrigatório)

Compare manifests e lockfiles com a base. Toda dependência **nova** ou **major bump** exige aprovação explícita (descrição, ADR, commit, instrução do projeto ou confirmação do usuário).

| Situação | Classificação padrão |
|---|---|
| Nova dependência sem aprovação demonstrada | 🔴 ALTO |
| Com advisory conhecido, licença incompatível ou substituível por lib já presente | 🚫 IMPEDITIVO |
| Bump minor/patch coerente com o projeto | Registre em cobertura; sem achado |

**Não afirme** comportamento ou advisory de memória. Use fonte oficial com URL real ou marque como `não verificado`.
