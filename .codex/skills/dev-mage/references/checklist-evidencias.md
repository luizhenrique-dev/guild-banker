# Matriz critério → evidência

Monte uma linha por critério **antes** de implementar.

| Critério | Evidência | Resultado esperado | Execução | Localização | Resultado |
|---|---|---|---|---|---|
| `T3-AC1` comportamento | teste / inspeção / manual | asserção ou observação | `cmd` em `dir/` | `path/test_x.py::test_y` (planejado) | pendente |

Regras:
- Testes planejados são marcados como `(planejado)` até existirem.
- A matriz pode evoluir com justificativa, preservando cobertura; critérios não mudam.
- Critério vago que altere comportamento esperado → bloqueio.
- Critério manual → procedimento escrito + responsável pela aprovação. Sem autoaprovação.
- Quando o critério exige integração, mock isolado não é evidência.

## Validar uma evidência

Antes de marcar `aprovado`, confirme na saída real:
1. Comando executado, diretório e resultado.
2. Os testes relevantes foram **descobertos e executados** (não só "processo terminou").
3. Skips, filtros, timeouts, avisos relevantes.
4. As asserções observam o comportamento do critério.
5. Contagens e limitações só se lidas na saída.

Resultados: `pendente` · `aprovado` · `falhou` · `impedido` (necessário, não executável aqui).

## Validação por momento

| Momento | Escopo |
|---|---|
| Durante a implementação | Menor teste que diagnostica a alteração |
| Fechar task | Evidências dos critérios + gates obrigatórios dos componentes afetados |
| Fechar spec | Verificação integrada da spec/projeto |

Não imponha ordem universal entre lint/build/test; respeite as dependências dos comandos do projeto. Não reutilize resultado anterior se alterações posteriores puderem invalidá-lo.
