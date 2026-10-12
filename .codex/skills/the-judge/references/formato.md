# Formato do achado e template de relatório

## Formato do achado (§ por achado)

IDs estáveis P1, P2… Ordene por risco (🚫 → 🔴 → 🟡 → 🟢 → 💡 → ❓ → 👏).

```
P1 — 🔴 ALTO — <título específico> [🔍 REVISÃO MANUAL, se aplicável]
Local: arquivo:linha ou intervalo, com versão quando necessário.
Problema: comportamento e condição que o provoca.
Impacto: consequência concreta.
Evidência: caminho de execução, teste, resultado observado ou fonte verificada com URL real.
Solução: menor mudança adequada.
Justificativa: por que resolve e qual trade-off relevante existe.
Validação: cenário de teste ou check necessário (quando aplicável).
```

**Variações por categoria:**
- **❓ DÚVIDA** — substitua Problema/Impacto por **Contexto** e **Pergunta**
- **💡 SUGESTÃO** — use **Valor agregado** em vez de Impacto
- **👏 ELOGIO** — apenas Local e descrição concreta do que foi bem feito
- **🔍 REVISÃO MANUAL** — indique especificamente o que o revisor humano deve verificar

Critique o código, não a pessoa. Não apresente pergunta como defeito.

---

## Template de relatório (uma única resposta)

### 1. Resumo
O que mudou; veredito e motivo principal; base/ancestral/HEAD e escopo incluído/excluído; limitações que afetam a conclusão.

### 2. Escopo e dependências
Tabela: arquivo/dep · relação com a feature (dentro/fora) · aprovação encontrada (sim/não/onde).
**Sempre presente**, mesmo que "nenhum desvio detectado."

### 3. Achados
Por risco, formato acima. Se vazio: *"Nenhum achado confirmado no escopo revisado."*

### 4. Revisão manual recomendada
Lista dos 🔍 com o que verificar. Omita se vazio.

### 5. Checks
Comando, resultado, snapshot, contagens, limitações; gaps de teste com cenário esperado.

### 6. Dúvidas
Só as necessárias; indique quais impedem o merge.

### 7. Pontos positivos
Apenas 👏 concretos; não equilibre artificialmente.

### 8. Cobertura
Revisado, excluído, fontes externas (URL real + o que sustenta), preexistentes relevantes.

### 9. Próximos passos
- Corrigir antes do merge
- Acompanhar depois
- Opcional
- Candidatos a automação (lint/teste/CI) quando úteis

**Regras gerais:**
- Omita seções opcionais vazias.
- Não reproduza grandes blocos de código sem necessidade.
- Contagens e limitações só se lidas na saída real.
- Nunca diga "zero bugs" ou "código perfeito".
- Título do relatório compatível com o veredito real.
