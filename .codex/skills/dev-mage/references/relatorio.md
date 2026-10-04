# Templates de saída

## Inventário (fim da Fase 0)

```
## Inventário
- Spec: <path> | Escopo: spec | task <id> | resume
- Elegíveis: T1, T3 | Dependentes: T4(T3) | Bloqueadas: —
- Comandos confirmados: `make test` (raiz) · `npm run lint` (web/)
- Gates: lint obrigatório · e2e impedido (sem Docker)
- Baseline (módulos tocados): 48 passed, 2 failed preexistentes (test_x, test_y)
- Working tree: limpo | alterações preexistentes em <paths> (não tocadas)
- Limitações: sem delegação · sem rede
```

## Fechamento de task (conciso)

```
### T3 — concluida | validada, commit pendente | bloqueada
- Critérios: T3-AC1 ✅ test_a · T3-AC2 ✅ test_b · T3-AC3 ⛔ impedido (e2e)
- Validações: `make test` → 51 passed, 2 failed preexistentes · `npm run lint` → ok
- Revisão: própria (sem independente) | independente: <agente>
- Git: abc1234 | pendente — <motivo>
- Decisões: <id> <uma linha> | —
```

## Relatório final

```
## DevMage — <concluída | parcial | bloqueada>
**Escopo:** <spec|task>
**Tasks:** concluídas T1, T3 · pendentes T4 · bloqueadas T2 (<motivo>)
**Critérios:** <n> com evidência · não validados: <lista + motivo>
**Validações:** <comando → resultado observado, por escopo> · limitações: <...>
**Revisão:** <independente|própria> para <tasks de alto risco>
**Git:** commits abc1234, def5678 · pendente: <task — motivo> · residual: <paths|nenhum>
**Decisões:** <lista ou —>
**Próximas ações:** <só ações reais: verificações manuais, decisões pendentes, push/PR pelo usuário>
```

Regras: título compatível com o estado real; contagens só se lidas na saída; nunca "zero regressões"; não declare conclusão total com critério, validação ou aprovação em aberto.
