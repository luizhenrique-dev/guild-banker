# Contrato esperado da spec do dev-quest

Ajuste este arquivo à estrutura real produzida pelo `dev-quest`. Se a spec encontrada divergir, siga a spec e registre a divergência no relatório.

## Spec principal (`<spec>/README.md` ou arquivo único)

Seções esperadas:
- **Objetivo e escopo** — o que entra e o que fica fora.
- **Restrições** — técnicas, de arquitetura, de dependências.
- **Critérios de aceite globais** — identificados (ex.: `AC-G1`).
- **Contratos compartilhados** — interfaces, schemas, eventos usados por mais de uma task.
- **Índice de tasks** — id, título, status, dependências, arquivo.
- **Verificação final** — comandos/procedimentos de fechamento da spec.

## Arquivo de task (`<spec>/tasks/T<n>-<slug>.md`)

- Referência à spec principal.
- Objetivo da task e arquivos previstos.
- Dependências (ids de tasks).
- Critérios de aceite da task (ex.: `T3-AC1`), cada um testável.
- Seção **Status** com estado, e seção **Evidências** (preenchida pelo dev-mage).

## Estados

`pendente` → `em_andamento` → `concluida` | `bloqueada`

- Registre o estado no próprio arquivo da task **e** no índice da spec principal, se existir. Não crie fontes de status adicionais.
- `concluida` exige: todos os critérios obrigatórios com evidência observada + gates obrigatórios aprovados. Commit é registrado à parte.

## Onde registrar

| Informação | Local |
|---|---|
| Estado da task | Arquivo da task + índice |
| Matriz critério→evidência | Seção Evidências da task |
| Decisões aprovadas | Seção Decisões da spec principal (criar se não houver, uma só) |
| Bloqueios | Arquivo da task + índice |
