# MJV challenge

Base de um sistema assíncrono em Go, com API HTTP, worker, RabbitMQ e MongoDB.

## Executar

```bash
docker compose up --build
```

Serviços expostos:

- API: `http://localhost:8080`
- RabbitMQ Management: `http://localhost:15672` (`app` / `app`)
- MongoDB: `mongodb://root:root@localhost:27017/?authSource=admin`

## Fluxo implementado

1. `POST /jobs` recebe um job e o publica na fila durável `jobs`.
2. O worker consome a mensagem com confirmação manual.
3. Após persistir o job na collection `jobs`, o worker confirma a mensagem.
4. Em falhas, a mensagem volta para a fila para uma nova tentativa.

```bash
curl -i -X POST http://localhost:8080/jobs \
  -H 'Content-Type: application/json' \
  -d '{"type":"send-email","payload":{"recipient":"ana@example.com"}}'
```

O endpoint `GET /health` retorna `204 No Content`.

## Organização

`core` contém o domínio, as portas e os casos de uso. Cada integração possui
`adapter`, `dto`/`model` e `mapper`, isolando HTTP, RabbitMQ e MongoDB.
