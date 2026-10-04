---
name: grill-me
description: Entrevista implacável para afiar um plano ou design, expondo lacunas, contradições e ambiguidades, e consolidando uma proposta final.
author: "Luiz Henrique FS"
---

<!-- Adaptado da skill "grilling" de https://github.com/mattpocock/skills (MIT License, Matt Pocock). -->

Entreviste o usuário até chegar a um entendimento compartilhado. Modele o problema como uma **árvore de design**: cada decisão ramifica nas
decisões que dependem dela.

## 0. Diagnóstico inicial

Antes do primeiro round, leia todo o contexto disponível (pedido, documentos, código, decisões anteriores) e liste explicitamente:

- **Buracos**: o que precisa ser decidido e não foi mencionado.
- **Contradições**: requisitos, restrições ou premissas que se anulam.
- **Ambiguidades**: termos ou requisitos que admitem mais de uma leitura relevante.
- **Faltas**: informações necessárias para a proposta que ninguém forneceu.

Cada item vira um nó da árvore. Contradições têm prioridade: não faz sentido detalhar ramos que dependem de algo inconsistente.

Calibre a profundidade ao tamanho do problema. Um problema pequeno pode se resolver em um round; não fabrique perguntas para parecer rigoroso.

## 1. Rounds e frontier

Trabalhe a árvore em **rounds**. A **frontier** é toda decisão cujos pré-requisitos já estão resolvidos: as perguntas que você pode fazer _agora_
sem adivinhar respostas que ainda não ouviu. Pergunte toda a frontier em um round: numere cada pergunta e dê sua resposta recomendada. Depois
aguarde as respostas antes do próximo round.

Formato de um round:

```
❓ **Q1** - **<título>**: <corpo da pergunta, pode ter vários parágrafos e múltipla escolha>

➡️ <sua resposta recomendada, com o motivo ancorado no contexto/projeto>

---

❓ **Q2** - **<título>**: <corpo>

➡️ <resposta recomendada>
```

Cada round reformula a árvore: decisões resolvidas empurram a frontier e desbloqueiam perguntas dependentes. Recalcule a frontier e faça o próximo
round. Uma pergunta cuja resposta depende de outra ainda aberta neste round pertence a um round _posterior_.

Se uma resposta do usuário contradiz algo dito antes (ou algo observado no ambiente), aponte a contradição explicitamente no round seguinte e
peça a resolução antes de avançar nos ramos afetados.

## 2. Decisões vs. premissas

Nem todo nó da árvore precisa virar pergunta. Separe:

- **Decisões**: escolhas de negócio, trade-offs ou preferências que só o usuário pode fazer. Pergunte e aguarde.
- **Premissas**: pontos em aberto onde o contexto e o projeto apontam claramente para uma resolução. Registre como
  `⚠️ Assumindo: <inferência> — porque <evidência no contexto>` e siga em frente; o usuário pode contestar a qualquer momento.

Toda inferência deve citar em que se baseia (código existente, convenção do projeto, restrição declarada). Inferência sem base vira pergunta.

## 3. Fatos são sua responsabilidade

Descobrir _fatos_ é seu trabalho, nunca do usuário. Quando uma pergunta da frontier precisa de um fato do ambiente (filesystem, ferramentas, etc.),
despache um sub-agente para buscá-lo; não pergunte ao usuário nada que você poderia verificar. Não bloqueie: uma exploração em andamento é um
pré-requisito não resolvido, então só as perguntas a jusante esperam o sub-agente; pergunte o resto da frontier agora. As _decisões_ são do usuário.

## 4. Encerramento e proposta consolidada

A entrevista termina quando a frontier está vazia: todo ramo visitado, nada assumido em silêncio. Então entregue a **proposta consolidada**:

1. **Problema e necessidade** — reformulados com o entendimento final.
2. **Decisões tomadas** — cada uma com a justificativa dada pelo usuário.
3. **Premissas assumidas** — cada inferência com sua base; o usuário valida ou corrige.
4. **Contradições resolvidas** — o que conflitava e como foi resolvido.
5. **Riscos e pontos em aberto residuais** — o que não pôde ser fechado e a recomendação para cada um.
6. **Plano/design resultante** — a proposta em si, completa e assertiva.

Não aja sobre a proposta até o usuário confirmar que ela reflete o entendimento compartilhado.
