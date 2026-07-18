---
name: "GuildBanker"
colors:
  primary: "#0052ff"
  primary-active: "#003ecc"
  primary-disabled: "#a8b8cc"
  canvas: "#ffffff"
  surface-soft: "#f7f7f7"
  surface-strong: "#eef0f3"
  surface-dark: "#0a0b0d"
  surface-dark-elevated: "#16181c"
  hairline: "#dee1e6"
  ink: "#0a0b0d"
  body: "#5b616e"
  muted: "#7c828a"
  on-primary: "#ffffff"
  semantic-up: "#05b169"
  semantic-down: "#cf202f"
typography:
  display:
    fontFamily: Inter
    fontSize: 3.25rem
    fontWeight: 400
  heading:
    fontFamily: Inter
    fontSize: 2rem
    fontWeight: 600
  body:
    fontFamily: Inter
    fontSize: 1rem
    fontWeight: 400
  label:
    fontFamily: Inter
    fontSize: 0.875rem
    fontWeight: 500
  mono:
    fontFamily: "JetBrains Mono"
    fontSize: 1rem
    fontWeight: 500
spacing:
  xs: "8px"
  sm: "12px"
  md: "20px"
  lg: "24px"
  xl: "32px"
rounded:
  sm: "8px"
  md: "12px"
  lg: "24px"
  full: "9999px"
  pill: "100px"
---

## Overview
GuildBanker é uma plataforma de controle financeiro pessoal e familiar. O design segue o estilo institucional-calmo da Coinbase: canvas branco, tipografia editorial (Inter), azul único (#0052ff) reservado para CTAs primários e ênfase, e valores monetários sempre em fonte mono (JetBrains Mono). Público-alvo: famílias/grupos que compartilham controle de gastos. Intenção emocional: confiança institucional, clareza, calma.

## Color usage
- **#0052ff (Coinbase Blue)** é a única cor de ação: CTAs primários, wordmark, links de ênfase, badge de visibilidade PUBLIC. Usar com parcimônia — 1-2 momentos de azul por seção.
- **#ffffff** é o piso padrão; **#f7f7f7** e **#eef0f3** para bandas de elevação suave e fundos de botões secundários/inputs de busca.
- **#0a0b0d** para heroes escuros (ex.: hero do login) e texto principal.
- Verde (#05b169) e vermelho (#cf202f) são **semânticos apenas em texto** — nunca fundo de botão. Vermelho para o total família e valores de despesa; verde para receitas.

## Typography
Inter para tudo (display 400 com tracking negativo, body 400/600). Números monetários e tabulares sempre em JetBrains Mono weight 500. Hierarquia por peso/tamanho, não por cor.

## Layout
App logado: topbar 64px + sidebar de navegação + conteúdo. Marketing/login: canvas branco editorial. Cards com radius 24px, inputs 12px, botões pill 100px, avatares circulares. Densidade maior atrás do login; generosidade editorial no login/onboarding.

## Do's and Don'ts
- Do: reservar o azul para CTAs primários e ênfase; renderizar todo valor em JetBrains Mono; usar badges de visibilidade distintos (cadeado cinza para PRIVATE, pessoas azul para PUBLIC).
- Do: usar categorias apenas do enum fixo com labels pt-BR.
- Don't: introduzir uma segunda cor de marca; usar verde/vermelho como fundo de botão; aplicar sombras em excesso (um único tier suave); cantos retos em CTAs.
