# MJV challenge

Sistema assíncrono de pedidos em Go. A API cria pedidos por HTTP e o worker os
processa de forma assíncrona com MongoDB, RabbitMQ e transactional outbox.

Os detalhes de desenho, fluxo de dados, resiliência, mensageria e telemetria
estão em [ARCHITECTURE.md](ARCHITECTURE.md).

## Serviços e acessos locais

O [docker-compose.yml](docker-compose.yml) inicia somente a aplicação e suas
dependências. As interfaces de desenvolvimento ficam no
[docker-compose.interfaces.yml](docker-compose.interfaces.yml), não exigem
arquivo `.env` e só são iniciadas quando solicitadas.

| Serviço | Compose | Porta | Responsabilidade |
| --- | --- | --- | --- |
| `api` | base | `8080` | Expõe HTTP, Swagger e cria pedidos. |
| `worker` | base | — | Publica a outbox e processa pedidos. |
| `rabbitmq` | base | interfaces: `5672`, `15672` | Broker e painel de filas. |
| `mongo` | base | interfaces: `27017` | Armazena pedidos e eventos de outbox. |
| `mongo-init` | base | — | Inicializa replica set e índices. |
| `mongo-express` | interfaces | `8081` | Visualização dos documentos. |
| `jaeger` | interfaces | `16686` | Visualização de traces e métricas RED. |
| `prometheus` | interfaces | `9090` | Armazena métricas RED derivadas dos spans. |

Com as interfaces ativas, os acessos são:

- API: `http://localhost:8080`
- Swagger UI: `http://localhost:8080/swagger/index.html`
- RabbitMQ Management: `http://localhost:15672` (`app` / `app`)
- Mongo Express: `http://localhost:8081` (`admin` / `admin`)
- Jaeger UI: `http://localhost:16686`
- Prometheus: `http://localhost:9090`

Após gerar pedidos, abra a aba **Monitor** no Jaeger para visualizar taxa de
requisições, erros e percentis de duração por serviço e operação. A primeira
agregação pode levar até um minuto.

## Executar e operar

O Makefile é um atalho para os comandos Docker Compose equivalentes abaixo.
Na primeira execução local sem Docker, execute `make deps` para baixar os
módulos Go.

| Ação | Makefile | Docker Compose |
| --- | --- | --- |
| Construir imagens | `make build` | `docker compose -f docker-compose.yml build` |
| Subir stack base | `make up` | `docker compose -f docker-compose.yml up -d --scale worker=1` |
| Reconstruir e subir | `make rebuild` | `docker compose -f docker-compose.yml up --build -d --scale worker=1` |
| Subir com interfaces | `make up-interfaces` | `docker compose -f docker-compose.yml -f docker-compose.interfaces.yml up -d --scale worker=1` |
| Escalar workers | `make scale-workers WORKER_REPLICAS=4` | `docker compose -f docker-compose.yml up -d --scale worker=4 worker` |
| Acompanhar logs | `make logs` | `docker compose -f docker-compose.yml -f docker-compose.interfaces.yml logs -f` |
| Listar containers | `make ps` | `docker compose -f docker-compose.yml -f docker-compose.interfaces.yml ps` |
| Parar stack | `make down` | `docker compose -f docker-compose.yml -f docker-compose.interfaces.yml down` |
| Regenerar Swagger | `make swagger` | `go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g main.go -d cmd/api,internal/core/error,internal/inbound/http/adapter,internal/inbound/http/dto -o docs` |

Para iniciar quatro workers já no primeiro comando, use
`make up WORKER_REPLICAS=4` ou substitua `worker=1` por `worker=4` no comando
Docker Compose. A API mantém uma única réplica porque publica diretamente a
porta `8080` no host.

Para publicar a API em outro ambiente, altere `SWAGGER_HOST` no serviço `api`
do Compose para o host público, sem protocolo.

## API HTTP

| Método e rota | Resposta | Descrição |
| --- | --- | --- |
| `POST /orders` | `201` | Cria o pedido e registra seu processamento assíncrono. |
| `GET /orders/:id` | `200` / `404` | Busca um pedido pelo identificador. |
| `GET /health` | `204` | Healthcheck leve da API. |
| `GET /swagger/index.html` | `200` | Interface Swagger. |

Exemplo de criação:

```bash
curl -i -X POST http://localhost:8080/orders \
  -H 'Content-Type: application/json' \
  -d '{"product_name":"Caderno","quantity":2}'
```

`product_name` é obrigatório e `quantity` deve ser maior que zero. Violações
retornam `400` antes de qualquer persistência, com `err` e, quando aplicável,
uma lista `errors`:

```json
{"err":"invalid request data","errors":[{"field":"quantity","message":"must be greater than zero"}]}
```

O documento retornado contém `id`, `status`, `created_at` e `updated_at`.

## Testes

Execute a suíte unitária:

```bash
go test ./...
```

Para medir a cobertura do código unitariamente testável:

```bash
make coverage
```

Os testes automatizados são unitários. A integração com MongoDB e RabbitMQ pode
ser validada ao subir o Compose e pelo procedimento manual de DLQ em
[ARCHITECTURE.md](ARCHITECTURE.md#testar-a-dlq-manualmente).
O alvo de cobertura exclui drivers, adapters de integração, pontos de
composição em `cmd/` e Swagger gerado; o mínimo configurado é 80%.

## Teste de carga

O cenário em [tests/load/orders.js](tests/load/orders.js) usa k6 por meio de
[docker-compose.load.yml](docker-compose.load.yml). Ele cria pedidos e consulta
cada `id` até `PROCESSADO`, medindo separadamente a aceitação HTTP e o tempo de
processamento assíncrono.

Com Makefile:

```bash
make up
make load-smoke
make load-sustainable WORKER_REPLICAS=4
make load-saturation WORKER_REPLICAS=4
make load-stress WORKER_REPLICAS=4 STRESS_PEAK=2000
```

Com Docker Compose, suba o stack antes e execute o perfil desejado:

```bash
docker compose -f docker-compose.yml up -d --scale worker=4 worker
COMPOSE_IGNORE_ORPHANS=true docker compose -f docker-compose.yml -f docker-compose.load.yml run --rm k6 run -e K6_PROFILE=stress -e K6_STRESS_PEAK=2000 -e WORKER_REPLICAS=4 /scripts/orders.js
```

Substitua `stress` por `smoke`, `sustainable` ou `saturation`; somente o perfil
`stress` usa `K6_STRESS_PEAK`. Os perfis executam um aquecimento antes da carga.
`sustainable` gera uma requisição a cada três segundos por réplica;
`saturation`, duas por segundo por réplica; e `stress` sobe de 100 até o pico
configurado, formando backlog intencionalmente.

Os pedidos de carga ficam em `orders` ou `outbox`. Execute-os apenas em
ambiente local de teste. Para zerar o ambiente, pare o stack e remova os
volumes:

```bash
docker compose -f docker-compose.yml -f docker-compose.interfaces.yml down -v --remove-orphans
```
