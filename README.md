# Microservice Subscriptions Service

Microservicio de SEMS para administrar planes y ciclo de vida de suscripciones.

## Arquitectura

- Go + Clean Architecture + DDD
- Capas: `domain`, `application`, `infrastructure`, `interfaces`
- Persistencia: PostgreSQL con GORM
- Integraciones: Stripe + Kafka

## Configuracion centralizada (Config Service)

Desde esta version, la configuracion compartida entre microservicios se obtiene del Config Service usando:

- `GET /api/v1/config/services`
- `GET /api/v1/config/kafka`
- `GET /api/v1/config/{service-name}`

El servicio consulta esos endpoints al iniciar cuando existe `CONFIG_SERVICE_URL`.

### Que sigue en variables de entorno (local/deploy)

- `SERVER_PORT` (o `PORT` en plataformas como Azure/Render)
- `SERVICE_NAME`
- `CONFIG_SERVICE_URL`
- `CONFIG_SERVICE_TIMEOUT_MS`
- `DATABASE_URL` o `POSTGRES_*`
- `STRIPE_SECRET_KEY`
- `STRIPE_PUBLISHABLE_KEY`
- `STRIPE_WEBHOOK_SECRET`
- `KAFKA_USERNAME` (si aplica)
- `KAFKA_PASSWORD` (si aplica)
- `KAFKA_CA_CERT` / `KAFKA_CA_CERT_PATH` (si aplica)

### Que ahora puede venir desde Config Service

- Kafka shared config (`kafkaEnabled`, brokers, protocol, mechanism, clientId)
- Nombres de topics Kafka
- Valores funcionales no secretos (ej. currency, Stripe price IDs)
- Otros defaults operativos compartidos

Notas:

- Variables locales tienen prioridad sobre Config Service (override por entorno).
- Si Config Service no responde, se usan fallbacks existentes para no romper arranque.

## Endpoints

### Subscription Plans

- `GET /api/v1/subscription-plans`
- `GET /api/v1/subscription-plans/:planId`
- `POST /api/v1/subscription-plans`
- `PUT /api/v1/subscription-plans/:planId`
- `PATCH /api/v1/subscription-plans/:planId/deactivate`

### Subscriptions

- `GET /api/v1/subscriptions/:subscriptionId`
- `GET /api/v1/subscriptions/users/:userId`
- `POST /api/v1/subscriptions`
- `PATCH /api/v1/subscriptions/:subscriptionId/cancel`
- `PATCH /api/v1/subscriptions/:subscriptionId/change-plan`

### Webhooks

- `POST /api/v1/webhooks/stripe`

### Health

- `GET /health`

## Ejecutar localmente

1. Copia `.env.example` a `.env` y completa secretos.
2. Asegura que `CONFIG_SERVICE_URL` apunte a tu Config Service local.
3. Ejecuta:

```bash
go mod tidy
go run .
```

Por defecto corre en `http://localhost:8081`.

## Azure Container Apps

Recomendacion de variables en ACA:

- No secret env vars:
  - `SERVICE_NAME=subscriptions-service`
  - `CONFIG_SERVICE_URL=https://<config-service-domain>`
  - `CONFIG_SERVICE_TIMEOUT_MS=3000`
- Secret env vars (ACA Secrets + env refs):
  - `DATABASE_URL` o `POSTGRES_PASSWORD`
  - `STRIPE_SECRET_KEY`
  - `STRIPE_WEBHOOK_SECRET`
  - `KAFKA_PASSWORD`
  - `KAFKA_CA_CERT` (si aplica)

Sugerencias de despliegue:

- Configurar health probe sobre `GET /health`.
- Mantener `PORT` inyectado por ACA (la app ya lo prioriza sobre `SERVER_PORT`).
- Centralizar en Config Service los datos compartidos para evitar drift entre microservicios.

## Notas funcionales

- `user_id` proviene de IAM y se persiste sin FK cruzada.
- Se mantiene contrato actual con API Gateway y rutas existentes.
- No existen tablas de `payments`, `invoices`, `transactions` o `billing` en este servicio.
