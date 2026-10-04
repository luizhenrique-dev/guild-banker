# Protocolo de bloqueio

## Bloqueie quando houver
- Ambiguidade que muda comportamento observável ou contrato.
- Incompatibilidade real entre requisitos e restrições aplicáveis.
- Informação indispensável não obtida após investigação local razoável.
- Decisão fora do escopo, irreversível ou de impacto relevante.
- Gate obrigatório falhando ou impedido, sem resolução no escopo.
- Necessidade de reduzir cobertura ou alterar expectativas sem respaldo no contrato.
- Risco de sobrescrever, descartar ou incorporar alterações alheias.

## Não bloqueie por
- Escolha local reversível sustentada pelos padrões do projeto.
- Diferença intencional entre código atual e spec.
- Ausência justificada de ferramenta não obrigatória.

## Template

```
### Bloqueio — <task-id>
**Problema:** <descrição objetiva>
**Evidência:** <arquivo:linha, trecho ou erro — sem segredos>
**Impacto:** <o que fica impedido; tasks afetadas>
**Proposta:** <resolução recomendada + justificativa> (alternativa só se real e relevante)
**Confirmação necessária:** <pergunta direta ou recurso necessário>
```

## Após a resposta
Registre como **decisão aprovada** ou **informação do usuário**; verifique afirmações técnicas quando necessário; atualize requisitos/evidências só conforme a decisão; retome de um estado consistente.
