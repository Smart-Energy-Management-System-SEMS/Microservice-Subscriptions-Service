# Microservice Subscriptions Service

Microservicio de SEMS para administrar planes y ciclo de vida de suscripciones.

## Health endpoint

- `GET /api/v1/health`
- `GET /health`

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
- Localmente con Kafka en Docker usa `KAFKA_BROKERS=localhost:9092`.
- El servicio acepta `PORT` (y mantiene compatibilidad con `SERVER_PORT`).

## Build de contenedor

```bash
docker build -t subscriptions-service:local .
```

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

## Ejemplo local (sin contenedor)

```bash
# .env (local)
PORT=8083
CONFIG_SERVICE_URL=http://localhost:8090
KAFKA_BROKERS=localhost:9092
KAFKA_SECURITY_PROTOCOL=PLAINTEXT
DATABASE_URL=postgresql://user:password@localhost:5432/subscriptions?sslmode=disable
GIN_MODE=release

go run .
```

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
    KAFKA_BROKERS=<kafka-broker1:9093,kafka-broker2:9093> \
    KAFKA_SECURITY_PROTOCOL=SASL_SSL \
    KAFKA_SASL_MECHANISM=PLAIN \
    KAFKA_USERNAME=<kafka-username> \
    KAFKA_PASSWORD=<kafka-password> \
    DATABASE_URL="<postgres-connection-string>"
```

Recomendado en Azure:
- Definir `KAFKA_PASSWORD`, `DATABASE_URL` y secretos Stripe como secretos de Container Apps.
- Referenciar secretos con `secretref:` en `--env-vars` cuando aplique.
