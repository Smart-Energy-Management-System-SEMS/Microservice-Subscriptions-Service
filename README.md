# Microservice Subscriptions Service

Microservicio de SEMS para administrar planes y ciclo de vida de suscripciones.

## Local quickstart

- Config-Service local: `http://localhost:8090`
- API Gateway local: `http://localhost:8081`
- Puerto local de este microservicio: `8083`
- Base URL local final: `http://localhost:8083`

## Variables de entorno

### No sensibles

- `SERVER_PORT=8082`
- `SERVICE_NAME=subscriptions-service`
- `CONFIG_SERVICE_URL=http://localhost:8090`
- `CONFIG_SERVICE_TIMEOUT_MS=3000`

### Sensibles

- `DATABASE_URL` o `POSTGRES_*` (incluye password)
- `STRIPE_SECRET_KEY`
- `STRIPE_PUBLISHABLE_KEY`
- `STRIPE_WEBHOOK_SECRET`
- `KAFKA_PASSWORD` (si aplica)
- `KAFKA_CA_CERT` / `KAFKA_CA_CERT_PATH` (si aplica)

## Config-Service integration

Al iniciar, el servicio consulta:

- `GET /api/v1/config/services`
- `GET /api/v1/config/kafka`
- `GET /api/v1/config/{service-name}`

Con fallback local para no romper el arranque si Config-Service no responde.

## Health checks

Endpoints públicos sin autenticación:

- `GET /health`
- `GET /api/v1/health`

## Rutas reales expuestas

Route prefix: `/api/v1`

- `GET /api/v1/subscription-plans`
- `GET /api/v1/subscription-plans/:planId`
- `POST /api/v1/subscription-plans`
- `PUT /api/v1/subscription-plans/:planId`
- `PATCH /api/v1/subscription-plans/:planId/deactivate`
- `GET /api/v1/subscriptions/:subscriptionId`
- `GET /api/v1/subscriptions/users/:userId`
- `POST /api/v1/subscriptions`
- `PATCH /api/v1/subscriptions/:subscriptionId/cancel`
- `PATCH /api/v1/subscriptions/:subscriptionId/change-plan`
- `POST /api/v1/webhooks/stripe`

## CORS local

Permitidos:

- `http://localhost:3000`
- `http://localhost:5173`

Configuración activa:

- `Access-Control-Allow-Credentials: true`
- Methods: `GET, POST, PUT, PATCH, DELETE, OPTIONS`
- Headers: `Authorization, Content-Type, Origin, Accept, X-Requested-With`

## Auth/JWT

Este microservicio actualmente no aplica middleware JWT propio. Si en Gateway usas `API_GATEWAY_AUTH_REQUIRED=false`, se puede probar directo sin token.

Endpoints públicos recomendados en Gateway:

- `/health`
- `/api/v1/health`
- `/api/v1/webhooks/stripe`

## Dependencias locales

Kafka local (ya levantado):

```powershell
docker ps
```

PostgreSQL: asegúrate de que `DATABASE_URL` o `POSTGRES_*` apunten a una instancia accesible.

## Ejecutar local

```bash
go mod tidy
go run .
```
