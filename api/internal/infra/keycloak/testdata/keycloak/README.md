# Fixture fictícia emitida pelo Keycloak

JWT original emitido pelo Keycloak **26.7.5**, executado com OpenJDK21 em uma
distribuição standalone extraída em `/tmp/issue1-tools/keycloak-26.7.5`.
Captura registrada em `metadata.json` (UTC). Identidade exclusivamente fictícia:
`fixture-member`, UUID terminado em 42, role `member`. `admin` existe no realm,
mas não foi atribuído a essa identidade.

`access-token.jwt` foi copiado de `access_token` sem alterar os três segmentos.
`jwks.json` contém somente duas chaves públicas (assinatura RS256 e encriptação).
Nenhuma chave privada, refresh token ou credencial pessoal foi versionada.
O password de exemplo abaixo serve somente para recriar o realm isolado em memória.

## Replay regular

Na raiz, com Go compatível com `api/go.mod` e compilador C para race:

```sh
make api/test/integration
```

`TestVerifierKeycloakFixture` e `TestAuthKeycloakFixture` usam as chaves públicas,
issuer original e relógio privado de teste em `valid_at`; após `expires_at + 1`
a mesma fixture é rejeitada com 401. Issuer, audience, exp e assinatura continuam
validados. Os testes exigem os arquivos e não usam skips ou Keycloak/PostgreSQL
ativos. Provider discovery e rotação de JWKS possuem testes separados.

## Recaptura isolada

Requisitos: distribuição Keycloak 26.7.5, JDK 21 com truststore válido, curl,
Python 3 e porta 18880 livre. Use uma distribuição nova em `/tmp/issue1-tools`;
não aplique este import sobre o realm/volume pessoal. Execute os comandos de
captura na raiz do repositório. Download utilizado:
[release 26.7.5](https://github.com/keycloak/keycloak/releases/tag/26.7.5).

O transporte usa `127.0.0.1:18880`. `--hostname=http://localhost:8080` conserva
`http://localhost:8080/realms/guild-banker` como issuer do token. O banco
`dev-mem` elimina o realm e suas chaves privadas ao encerrar o processo.
Nenhuma migration da API é executada.

Salve o JSON abaixo em
`/tmp/issue1-tools/keycloak-26.7.5/data/import/guild-banker-realm.json`:

```json
{
  "realm": "guild-banker",
  "enabled": true,
  "sslRequired": "none",
  "defaultSignatureAlgorithm": "RS256",
  "accessTokenLifespan": 300,
  "roles": {"realm": [{"name": "member"}, {"name": "admin"}]},
  "clients": [
    {"clientId": "guild-banker-api", "enabled": true, "bearerOnly": true, "protocol": "openid-connect"},
    {
      "clientId": "fixture-issuer",
      "enabled": true,
      "publicClient": true,
      "directAccessGrantsEnabled": true,
      "standardFlowEnabled": false,
      "protocol": "openid-connect",
      "fullScopeAllowed": true,
      "defaultClientScopes": ["basic"],
      "optionalClientScopes": [],
      "protocolMappers": [
        {"name": "api-audience", "protocol": "openid-connect", "protocolMapper": "oidc-audience-mapper", "config": {"included.client.audience": "guild-banker-api", "access.token.claim": "true", "id.token.claim": "false"}},
        {"name": "username", "protocol": "openid-connect", "protocolMapper": "oidc-usermodel-property-mapper", "config": {"user.attribute": "username", "claim.name": "preferred_username", "jsonType.label": "String", "access.token.claim": "true", "id.token.claim": "false"}},
        {"name": "realm-roles", "protocol": "openid-connect", "protocolMapper": "oidc-usermodel-realm-role-mapper", "config": {"claim.name": "realm_access.roles", "jsonType.label": "String", "multivalued": "true", "access.token.claim": "true", "id.token.claim": "false"}}
      ]
    }
  ],
  "users": [
    {
      "id": "00000000-0000-4000-8000-000000000042",
      "username": "fixture-member",
      "firstName": "Fixture",
      "lastName": "Member",
      "email": "fixture-member@example.test",
      "emailVerified": true,
      "enabled": true,
      "realmRoles": ["member"],
      "requiredActions": [],
      "credentials": [{"type": "password", "value": "fixture-local-only", "temporary": false}]
    }
  ]
}
```

O scope `basic` inclui o mapper de `sub`, conforme o
[guia oficial do Keycloak](https://www.keycloak.org/docs/26.5.7/server_admin/index.html).
O mapper de audience inclui `guild-banker-api` somente no access token.
Username e realm roles possuem mappers explícitos; os demais scopes foram
omitidos para manter a identidade fictícia pequena.

Inicie em um terminal, com `JAVA_HOME` apontando para o JDK 21:

```sh
/tmp/issue1-tools/keycloak-26.7.5/bin/kc.sh start-dev \
  --http-host=127.0.0.1 --http-port=18880 \
  --hostname=http://localhost:8080 --db=dev-mem \
  --import-realm --log-console-color=false
```

Neste ambiente, o JRE Debian foi extraído somente em `/tmp`. Como seu symlink
de cacerts não tinha destino, um truststore público temporário foi criado
com keytool e fornecido por
`JAVA_OPTS_APPEND=-Djavax.net.ssl.trustStore=/tmp/issue1-tools/java/truststore.p12`.
Esse ajuste do runtime não é necessário em um JDK instalado com cacerts válido;
nenhum check de assinatura/issuer/audience da API foi alterado.
O sandbox precisou autorizar sockets para o processo isolado.

Aguarde a mensagem de bootstrap completo. Os seguintes requests foram validados:

```sh
curl -fL --silent --show-error \
  --output /tmp/issue1-tools/keycloak-discovery.json \
  http://127.0.0.1:18880/realms/guild-banker/.well-known/openid-configuration
```

Crie o formulário **sem newline final**, para não anexar uma quebra ao password:

```sh
python3 -c 'from pathlib import Path; Path("/tmp/issue1-tools/keycloak-token-form.txt").write_text("grant_type=password&client_id=fixture-issuer&username=fixture-member&password=fixture-local-only")'
curl -fL --silent --show-error \
  --header 'Content-Type: application/x-www-form-urlencoded' \
  --data-binary @/tmp/issue1-tools/keycloak-token-form.txt \
  --output /tmp/issue1-tools/keycloak-token-response.json \
  http://127.0.0.1:18880/realms/guild-banker/protocol/openid-connect/token
curl -fL --silent --show-error \
  --output /tmp/issue1-tools/keycloak-jwks.json \
  http://127.0.0.1:18880/realms/guild-banker/protocol/openid-connect/certs
```

Salve e execute o script abaixo na raiz. Ele copia apenas o access token e
JWKS público, confere algoritmo/kid/issuer/audience/sub/role e grava os metadados.
Não imprima o JSON da resposta nem copie refresh_token para o repositório.

```python
import base64
import datetime
import json
from pathlib import Path

scratch = Path('/tmp/issue1-tools')
fixture = Path('api/internal/infra/keycloak/testdata/keycloak')
response = json.loads((scratch / 'keycloak-token-response.json').read_text())
token = response['access_token']
decode = lambda part: json.loads(base64.urlsafe_b64decode(part + '=' * (-len(part) % 4)))
header, claims = map(decode, token.split('.')[:2])
jwks = json.loads((scratch / 'keycloak-jwks.json').read_text())
assert header['alg'] == 'RS256'
assert claims['iss'] == 'http://localhost:8080/realms/guild-banker'
assert 'guild-banker-api' in ([claims['aud']] if isinstance(claims['aud'], str) else claims['aud'])
assert claims['typ'] == 'Bearer'
assert claims['sub'] == '00000000-0000-4000-8000-000000000042'
assert claims['preferred_username'] == 'fixture-member'
assert 'member' in claims['realm_access']['roles']
assert 'admin' not in claims['realm_access']['roles']
assert any(key['kid'] == header['kid'] and key['alg'] == 'RS256' for key in jwks['keys'])
for key in jwks['keys']:
    assert not {'d', 'p', 'q', 'dp', 'dq', 'qi', 'oth', 'k'} & key.keys()
meta = {
    'keycloak_version': '26.7.5',
    'issuer': claims['iss'],
    'audience': 'guild-banker-api',
    'kid': header['kid'],
    'sub': claims['sub'],
    'preferred_username': claims['preferred_username'],
    'realm_roles': claims['realm_access']['roles'],
    'issued_at': claims['iat'],
    'expires_at': claims['exp'],
    'valid_at': claims['iat'] + 1,
    'captured_at': datetime.datetime.now(datetime.timezone.utc).isoformat(),
    'transport_url': 'http://127.0.0.1:18880',
    'identity': 'fictional local fixture; no production account',
}
fixture.mkdir(parents=True, exist_ok=True)
(fixture / 'access-token.jwt').write_text(token + '\n')
(fixture / 'jwks.json').write_text(json.dumps(jwks, indent=2) + '\n')
(fixture / 'metadata.json').write_text(json.dumps(meta, indent=2) + '\n')
print(json.dumps({'fixture_written': True, 'public_keys': len(jwks['keys']), 'algorithm': header['alg'], 'roles': meta['realm_roles']}))
```

Encerre somente esse terminal/processo com Ctrl-C. Na captura desta fixture,
o Keycloak confirmou shutdown; em seguida o alvo integration passou sem o
processo ativo. Recapturas geram novo kid/assinatura, preservando o mesmo contrato.
