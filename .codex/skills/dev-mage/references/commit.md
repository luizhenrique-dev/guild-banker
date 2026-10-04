# Commit local

Só se commits locais forem permitidos pelas instruções e aprovações do ambiente. Push e PR nunca.

## Sequência
1. Atualize status e evidências nos docs existentes.
2. Revise o diff completo da task, incluindo docs.
3. Faça staging **apenas** do atribuível à task (`git add -p` ou por caminho). Inspecione `git diff --cached`; nada alheio.
4. Commit único e coeso. Sem `add -A` indiscriminado, sem `--no-verify`.
5. Confirme hash e estado residual do working tree.
6. Se um hook alterou arquivos: revise, revalide o afetado, inclua ou reporte.

## Mensagem
- Convenção do repositório; na ausência, Conventional Commits.
- Inglês, baseada no diff preparado.
- Título ≤ 72 caracteres, imperativo, sem ponto final. Corpo com quebra em 72.
- Referencie o identificador de issue **fornecido** (título ou rodapé conforme convenção). Não invente ids.
- Breaking change comprovada → `!` no tipo e rodapé `BREAKING CHANGE:` descrevendo a incompatibilidade.

## Se não puder commitar ou falhar
Preserve as alterações e reporte separadamente:
- `implementação validada: sim|não`
- `commit: pendente — <motivo>`
Nunca declare commit realizado sem o hash observado.
