# Classificação de risco e veredito

## Tabela de risco

Cada achado recebe **exatamente uma** categoria. Categoria de domínio (ex.: "segurança") não define risco sozinha — impacto, alcance e condição real definem.

| Tag | Significado | Ação esperada |
|---|---|---|
| 🚫 **IMPEDITIVO** | Quebra o sistema, vulnerabilidade grave, perda de dados, pipeline inviabilizado, dependência com advisory/licença incompatível | Correção obrigatória antes de qualquer avanço |
| 🔴 **ALTO** | Estrutural severo: degradação a curto/médio prazo, N+1 sem paginação, falta de idempotência em fila/evento, violação de ADR, leak de recurso, mudança fora de escopo que aumenta risco, dependência nova sem aprovação | Correção obrigatória **ou** justificativa técnica formal aprovada pelo Tech Lead |
| 🟡 **MÉDIO** | Débito moderado: edge cases não tratados, acoplamento desnecessário, testes de integração ausentes, observabilidade insuficiente, escopo misturado sem risco direto | Ajuste no PR **ou** card de débito priorizado na sprint |
| 🟢 **BAIXO** | Legibilidade, simplificação pontual, otimização marginal sem efeito operacional | Opcional; nunca bloqueia com checks verdes. **Máx. 5** |
| 💡 **SUGESTÃO** | Evolução futura, padrão alternativo, candidato a ADR | Discussão; não impede aprovação |
| ❓ **DÚVIDA** | Intenção obscura, aparente divergência de requisito/arquitetura, hipóteses não confirmadas | Esclarecimento ou ajuste para tornar autoexplicativo |
| 👏 **ELOGIO** | Solução, teste ou aderência a padrão concretamente bem feitos | Nenhuma |

**🔍 REVISÃO MANUAL** — flag adicional (não substitui a categoria) para estruturas onde análise estática local é insuficiente e um humano deve validar com profundidade: concorrência/locks, transações distribuídas, migrations em tabelas volumosas, authn/authz, criptografia, cálculo financeiro, feature flags/rollout, infra/IaC, alterações de contrato público. Indique **o que** o revisor humano precisa verificar.

## Veredito

| Veredito | Condição |
|---|---|
| **SOLICITAR_ALTERAÇÕES** | Qualquer 🚫 ou 🔴 aberto, ou check obrigatório quebrado pelo delta |
| **INCONCLUSIVO** | Sem bloqueio confirmado, mas falta evidência essencial (❓ que impede concluir, check material não executado, 🔍 sem possibilidade de validação local) |
| **COMENTAR** | Sem impeditivo nem lacuna essencial; existem 🟡 |
| **APROVAR** | No máximo 🟢/💡/👏 e validação suficiente para o risco |

**Regras adicionais:**
- Limitação adicional não neutraliza impeditivo confirmado: SOLICITAR_ALTERAÇÕES + registro da limitação.
- Checks não executados não geram INCONCLUSIVO automaticamente; avalie materialidade.
- São recomendações locais, não aprovação no provedor Git.
