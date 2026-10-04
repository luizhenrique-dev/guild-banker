# Diagramas Mermaid

## Quando incluir
Siga a tabela de proporcionalidade do SKILL.md. Um diagrama só entra se tornar a solução mais verificável do que a prosa. Nunca invente telas, endpoints, eventos ou serviços; elementos novos levam marcação de proposto.

## Convenções
- Bloco de código com linguagem `mermaid`.
- Existente × proposto: use `classDef` (ex.: `classDef novo stroke-dasharray: 5 5`) ou sufixo `(proposto)` no rótulo.
- Rótulos nas setas indicam finalidade (`-->|publica evento|`).
- Identificadores técnicos exatos só se confirmados; caso contrário, nome conceitual.

## Arquitetural — `flowchart` com `subgraph`
Componentes · fronteiras aplicação/externos · dependências e comunicações rotuladas · existente vs. proposto.

## Fluxograma — `flowchart`
Gatilho · etapas · validações/decisões · caminho principal · alternativos e falhas · términos. Marque passos assíncronos; publicar evento ≠ processamento concluído.

## Sequência — `sequenceDiagram`
Atores/participantes · interações em ordem · respostas observáveis · `alt`/`opt`/`loop` para alternativas e erros.

## Validação
Se houver ferramenta no ambiente (ex.: `mmdc`), valide e registre. Caso contrário, escreva "sintaxe Mermaid não validada" na seção de rastreabilidade. Nunca afirme renderização que não observou.
