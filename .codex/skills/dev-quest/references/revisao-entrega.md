# Revisão e entrega

## Revisão em três perspectivas
- **Requisitos:** todo CA coberto por ≥1 task? Nenhum escopo não aprovado?
- **Engenharia:** caminhos verificados, contratos coerentes, dependências só para trás, estado válido ao fim de cada task?
- **Qualidade:** cada task tem critérios verificáveis e comandos confirmados? Testes incluídos na própria task?

Corrija antes de apresentar.

## Checklist
- [ ] ≤ 5 tasks; sem dependência circular ou futura
- [ ] 1 arquivo por task; links relativos corretos nos dois sentidos
- [ ] Índice da spec ≡ arquivos individuais
- [ ] Fatos × propostas × premissas × lacunas marcados
- [ ] Nenhum arquivo/comando/API não verificado apresentado como existente
- [ ] Mermaid: validado ou declarado não validado
- [ ] Nenhum código implementado
- [ ] Escopo e decisões da entrevista preservados

## Template de entrega

```
## DevQuest — <ISSUE> pronta para aprovação | bloqueada
**Solução:** <2–4 linhas>
**Arquivos:** gravados: <paths> | NÃO gravados (escrita bloqueada): <paths + conteúdo abaixo>
**Tasks:** T1 … · T2 … (dep T1) · … | CAs cobertos: CA1–CA<n>
**Premissas/riscos:** P1 … · P2 …
**Pendências não bloqueantes:** <item + sugestão>
**Próximo passo:** aprovar a spec → implementar com `dev-mage <path>` (autorização explícita necessária)
```

Peça aprovação explícita e aguarde. Não saia do plan mode nem inicie implementação.
