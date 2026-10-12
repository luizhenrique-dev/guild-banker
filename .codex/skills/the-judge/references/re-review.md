# Re-review e checklist final

## Re-review (com relatório anterior)

**Ao receber `--resume` ou um relatório anterior:**

1. Compare snapshots do relatório anterior com o estado atual; declare qualquer divergência.
2. Reavalie todos os achados abertos (não os feche sem evidência de correção).
3. Revise o delta novo e efeitos das correções sobre achados anteriores.
4. Preserve IDs originais (P1, P2…); não reabra achado cosmético em código inalterado.
5. Sem histórico verificável: declare que continuidade não foi verificável.

**Estados dos achados em re-review:**

| Estado | Significado |
|---|---|
| Resolvido | Correção observada com evidência |
| Aberto | Sem correção verificável |
| Contestação aceita | Achado retirado com justificativa técnica formal |
| Risco aceito | Decisão explícita registrada; não finja que o achado sumiu |
| Acompanhamento acordado | Sem inventar issue; registre o que foi combinado |

Impeditivo antigo descoberto tarde: declare a omissão e apresente evidência antes de incluir.

---

## Checklist final (executar antes de entregar o relatório)

- [ ] Base, snapshot e escopo explícitos; working tree tratado corretamente.
- [ ] Linhas citadas correspondem à versão revisada (não à HEAD se forem commits anteriores).
- [ ] Todo achado está ligado ao delta e tem evidência para cenário, impacto e solução.
- [ ] Passes G (escopo) e H (dependências) executados e reportados.
- [ ] Hipóteses estão em ❓, não em categorias de risco.
- [ ] 🔍 marcados onde análise local é insuficiente.
- [ ] Sem duplicação de tooling ou causa raiz agrupável; máx. 5 🟢.
- [ ] Checks e fontes são apenas os realmente executados/consultados.
- [ ] Veredito coerente com a tabela em `classificacao.md`.
- [ ] Nada foi alterado, instalado ou publicado durante o review.
