# Keycloak local — GuildBanker

Definições para Keycloak **26.7.5**, a versão do compose:

- `guild-banker-realm.json`: realm `guild-banker`, assinatura RS256, roles
  `member`/`admin` e usuário de teste Foo Bar (`foo.bar@email.com`).
- `guild-banker-clients.json`: importação parcial dos clients no realm existente.
  `guild-banker-ui` é público, com Authorization Code e PKCE S256 obrigatório.
  `guild-banker-api` é um recurso Bearer, sem login e sem exigir client secret na API.
- `prepare.py`: combina as duas definições em um único RealmRepresentation para
  importação de startup. A senha permanece como `${KC_TEST_USER_PASSWORD}`;
  é resolvida pelo Keycloak, não gravada pelo script.

O realm usa o e-mail como username, não permite editar o username separadamente
e bloqueia e-mails duplicados. Usernames são únicos por realm no Keycloak.

Foo Bar tem somente a role `member`, e-mail marcado como verificado e nenhuma
ação obrigatória no primeiro login para permitir testes automatizados locais.
Sua senha é definida por ambiente, sem credencial versionada. O UUID/sub é
`00000000-0000-4000-8000-000000000001`, estável para fixtures e vínculo com a API.

## Setup com Docker Compose

Requisitos: Docker/Compose, GNU Make, Python 3 e portas 5432/8080 disponíveis.
Na raiz do repositório:

```sh
cp infra/.env.example infra/.env
# Edite infra/.env e defina KC_TEST_USER_PASSWORD com sua senha local de teste.
make keycloak/validate
make keycloak/up
```

O alvo inicia Postgres e Keycloak com o overlay `docker-compose.import.yml`,
importa o JSON gerado em `infra/keycloak/import/` e mantém os volumes existentes.
Ele aguarda a prontidão dos serviços por até 180 segundos. O overlay consulta
`/health/ready` na porta de gerenciamento 9000, conforme o
[healthcheck do Keycloak](https://www.keycloak.org/observability/health).
`infra/.env` e o diretório gerado são ignorados pelo Git. Também é possível
exportar `KC_TEST_USER_PASSWORD` no terminal em vez de criar `infra/.env`.

O Keycloak fica em `http://localhost:8080`, admin local `admin`/`admin` conforme
o compose existente. Após o bootstrap, escolha `guild-banker` no console.
O hostname fixo conserva o issuer esperado pela API, inclusive se acessado
pelo transporte `127.0.0.1`. UI aceita apenas as origens locais 5173, com
redirects dentro dessas origens; adicione outras URLs explicitamente se precisar.

| Alvo | Efeito |
|---|---|
| `make keycloak/validate` | Valida os JSONs e os IDs, sem gerar arquivo ou iniciar serviços |
| `make keycloak/prepare` | Gera o arquivo combinado, sem Docker ou senha no host |
| `make keycloak/up` | Prepara e inicia o compose com importação inicial |
| `make infra/down` | Para os serviços sem excluir volumes |
| `make infra/logs` | Acompanha os logs dos serviços |

`make infra/up` continua usando o compose base, sem importação. Para fazer o
setup inicial do realm use `make keycloak/up`. Não é necessário usar `infra/clean`.

## Importação manual e realms existentes

Para uma instalação sem Compose, `make keycloak/prepare` gera um único JSON.
Exporte `KC_TEST_USER_PASSWORD` no ambiente do processo Keycloak e importe
`infra/keycloak/import/guild-banker-realm.json` usando o fluxo de startup
`--import-realm` (diretório `data/import`, arquivo `guild-banker-realm.json`).
O realm de origem também pode ser importado sozinho por startup/CLI com a senha
no ambiente do processo (não por upload do console); nesse caso importe os
clients separadamente pelo console, no realm `guild-banker` → **Partial import**,
selecionando `guild-banker-clients.json` e a opção **Skip** para recursos existentes.

Na [importação de startup](https://www.keycloak.org/server/importExport), realms
já existentes são ignorados. Portanto, repetir `keycloak/up` não atualiza nem
reseta clients, usuário ou senha. O arquivo de clients usa `ifResourceExists=SKIP`
na [importação parcial](https://www.keycloak.org/docs/latest/server_admin/index.html#_partial_import).
Para alterar configuração existente, edite o recurso no console ou planeje uma
atualização explícita; os alvos não usam overwrite ou remoção de realm.

## Tokens e API

O mapper do client UI inclui `aud: guild-banker-api` no access token. Os scopes
`basic`, `profile`, `email` e `roles` expõem `sub`, username e realm roles.
O ID token não ganha essa audience. Implicit Flow, password grant e service
accounts ficam desabilitados; testes de login usam Authorization Code com PKCE.
A SPA integra OIDC via keycloak-js; configurar as variáveis públicas conforme web/spa/.env.example.

Configuração correspondente em `api/config.env`:

```dotenv
KEYCLOAK_ISSUER=http://localhost:8080/realms/guild-banker
API_AUDIENCE=guild-banker-api
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://127.0.0.1:5173
```

A primeira chamada autenticada à API cadastra automaticamente `user_account`
com `external_id=sub`, se ausente; chamadas seguintes sincronizam apenas nome/e-mail
diferentes. O access token validado deve conter `name`/`email` não brancos e
`email_verified=true`. Perfil incompleto, colisão de e-mail com outro sub e
usuário local desativado recebem 403 genérico, sem associação ou reativação.
O perfil é snapshot do token: mudanças exigem emissão/renovação e tokens válidos
antigos podem reaplicar valores antigos; não há ordenação de versões. Cadastro
não concede membership nem garante autorização da rota.

Esses arquivos e o usuário são para desenvolvimento/teste local. Para implantação
fora desse ambiente, configure hostname HTTPS, redirects/origens da UI e
identidades/credenciais apropriadas ao ambiente.

## Validação realizada

Os JSONs foram importados em uma instância isolada Keycloak26.7.5 com banco
em memória. Foram confirmados: login de Foo Bar com a senha fornecida por env,
Authorization Code com PKCE S256, access token RS256 com issuer/sub/username/
e-mail/role member e audience da API, ID token sem audience da API, bloqueio de
PKCE ausente/redirect externo/password grant e reimportação parcial dos clients
com SKIP sem sobrescrita. A mesma sondagem Bash do overlay recebeu HTTP200 na
porta de gerenciamento da instância isolada. O processo foi encerrado após os testes.

`make keycloak/validate`, `make keycloak/prepare` e `docker compose config`
passaram; senha de ambiente ausente impede a configuração do overlay. A subida
com Docker não foi executada neste ambiente sem acesso ao daemon; os testes
de execução usaram a distribuição standalone da mesma versão.

## Logout da SPA
O client UI declara `post.logout.redirect.uris` com retornos exatos
`http://localhost:5173/` e `http://127.0.0.1:5173/` (separador `##` no JSON).
Em realms existentes, configure **Valid post logout redirect URIs** no client
`guild-banker-ui` com essas duas URLs pelo console; importação SKIP não atualiza
configurações existentes. Preserve as demais configurações. Não exclua o realm.
O SDK envia o ID token hint ao encerramento SSO e retorna à raiz da UI.

A suíte real é `npm --prefix web/spa run test:e2e:auth` na raiz. Ela lê a senha
do ambiente ou de infra/.env, desabilita artefatos de autenticação e não provisiona
cadastros por seed. Requer API atual com banco de teste preparado e perfil
completo/verificado no access token para a prova positiva de autorização. Refresh/expiração usam relógio virtual só no browser;
nenhum tempo de sessão do realm compartilhado é alterado.
