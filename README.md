# Microservice Subscriptions Service

Microservicio de SEMS para administrar planes y ciclo de vida de suscripciones.

## Health endpoint

- `GET /api/v1/health`
- `GET /health`

## Swagger

- `GET /swagger/index.html`
- `GET /swagger/doc.json`

Flujo recomendado para validar desde Swagger:
- Consultar los planes fijos con `GET /api/v1/subscription-plans`
- Crear una suscripcion con `POST /api/v1/subscriptions`
- Consultar con `GET /api/v1/subscriptions/{subscriptionId}`
- Probar cambio de plan o cancelacion con los `PATCH`

Nota:
- Este micro no crea planes por API. El catalogo permitido es fijo y se mantiene con `Free`, `Plus` y `Pro`.

## Entornos locales listos

- `.env`: Azure Event Hubs activo por defecto
- `.env.local.azure`: copia lista para Azure Event Hubs
- `.env.local.kafka`: copia lista para Kafka local en `localhost:9092`

## Variables requeridas

```env
PORT=8080
CONFIG_SERVICE_URL=
KAFKA_BROKERS=
KAFKA_SECURITY_PROTOCOL=
KAFKA_SASL_MECHANISM=
KAFKA_USERNAME=
KAFKA_PASSWORD=
DATABASE_URL=
GIN_MODE=release
```

Notas:
- En Azure no uses `localhost` para servicios externos.
- El servicio acepta `PORT` (y mantiene compatibilidad con `SERVER_PORT`).

## Build de contenedor

```bash
docker build -t subscriptions-service:local .
```

## Run con Docker Compose

```bash
docker compose up -d --build
```

Esto levanta:
- `subscriptions-service`
- `postgres`
- `kafka`

Valores por defecto importantes:
- app: `http://localhost:8083`
- postgres: `localhost:5432`
- kafka dentro de la red Docker: `kafka:9092`
- config-service externo al compose: `http://host.docker.internal:8090`

Si tu `config-service` corre local en tu máquina, el valor por defecto ya funciona desde el contenedor.
El compose usa variables con prefijo `COMPOSE_` para no chocar con tu `.env` local de `go run`.
Si quieres cambiar puertos o credenciales, puedes sobrescribir variables al ejecutar:

```bash
APP_PORT=8083 \
COMPOSE_CONFIG_SERVICE_URL=http://host.docker.internal:8090 \
POSTGRES_EXPOSE_PORT=5432 \
KAFKA_EXTERNAL_PORT=9092 \
docker compose up -d --build
```

Si prefieres que este micro no consulte el config-service, puedes dejar `COMPOSE_CONFIG_SERVICE_URL` vacío y pasar las variables necesarias por entorno.

## Run local con contenedor

```bash
docker run --rm -p 8080:8080 \
  -e PORT=8080 \
  -e CONFIG_SERVICE_URL=http://host.docker.internal:8090 \
  -e KAFKA_BROKERS=host.docker.internal:9092 \
  -e KAFKA_SECURITY_PROTOCOL=PLAINTEXT \
  -e DATABASE_URL="postgresql://user:password@host.docker.internal:5432/dbname?sslmode=disable" \
  subscriptions-service:local
```

Topic Kafka por defecto de este micro:
- `subscriptions.events`

Eventos publicados dentro del payload:
- `subscription.created`
- `subscription.cancelled`
- `subscription.plan.changed`
- `subscription.renewal.requested`
- `subscription.expired`
- `subscription.updated`

Aclaracion importante:
- El micro no publica `SubscriptionCreated` como nombre de evento en Kafka.
- El valor real del campo `eventType` es `subscription.created` y variantes equivalentes en minusculas para los otros eventos.

Formato del evento:

```json
{
  "eventType": "subscription.created",
  "eventId": "550e8400-e29b-41d4-a716-446655440000",
  "occurredAt": "2026-06-12T22:30:00Z",
  "data": {
    "subscriptionId": "sub-123",
    "userId": "user-123",
    "planId": "plan-456",
    "status": "ACTIVE",
    "requiresPayment": true,
    "amount": 15
  }
}
```

Notas de contrato:
- El topic fisico de publicacion es solo `subscriptions.events`.
- `eventType` identifica el tipo de evento dentro del envelope y no se usa como topic fisico.
- Si alguna documentacion previa mencionaba `SubscriptionCreated`, debe leerse como desactualizada para este servicio.

## Ejemplo Azure Container Apps

```bash
az containerapp create \
  --name subscriptions-service \
  --resource-group <RESOURCE_GROUP> \
  --environment <CONTAINERAPPS_ENV> \
  --image <ACR_LOGIN_SERVER>/subscriptions-service:latest \
  --target-port 8080 \
  --ingress external \
  --registry-server <ACR_LOGIN_SERVER> \
  --env-vars \
    PORT=8080 \
    GIN_MODE=release \
    CONFIG_SERVICE_URL=https://<config-service-url> \
    KAFKA_BROKERS=<namespace>.servicebus.windows.net:9093 \
    KAFKA_SECURITY_PROTOCOL=SASL_SSL \
    KAFKA_SASL_MECHANISM=PLAIN \
    KAFKA_USERNAME='$ConnectionString' \
    KAFKA_PASSWORD='Endpoint=sb://<namespace>.servicebus.windows.net/;SharedAccessKeyName=<policy>;SharedAccessKey=<key>' \
    KAFKA_TOPIC_SUBSCRIPTIONS_EVENTS=subscriptions.events \
    DATABASE_URL="<postgres-connection-string>"
```

Recomendado en Azure:
- Definir `KAFKA_PASSWORD`, `DATABASE_URL` y secretos Stripe como secretos de Container Apps.
- Referenciar secretos con `secretref:` en `--env-vars` cuando aplique.
