# Microservice Subscriptions Service

Microservice de SEMS para administrar planes de suscripcion, caracteristicas de planes y ciclo de vida de suscripciones. No maneja pagos directos.

## Arquitectura

- Go + Clean Architecture + DDD
- Capas: `domain`, `application`, `infrastructure`, `interfaces`
- Persistencia: PostgreSQL (Neon) con GORM
- Integraciones externas: Stripe (adapter), Kafka (publisher)
- API REST lista para exponer detras de API Gateway

## Variables de entorno

- `SERVER_PORT` (ej: `8081`)
- `DATABASE_URL` (opcional) o variables separadas:
- `POSTGRES_HOST`
- `POSTGRES_PORT`
- `POSTGRES_DB`
- `POSTGRES_USER`
- `POSTGRES_PASSWORD`
- `POSTGRES_SSLMODE`
- `POSTGRES_CHANNEL_BINDING`
- `STRIPE_SECRET_KEY`
- `STRIPE_WEBHOOK_SECRET`
- `KAFKA_BROKERS` (ej: `localhost:9092`)
- `KAFKA_CLIENT_ID` (ej: `subscriptions-service`)

Usa `.env.example` como plantilla y guarda tus secretos en `.env` (ignorado por Git).

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

## Estados de suscripcion

- `ACTIVE`
- `INACTIVE`
- `CANCELLED`
- `PENDING_RENEWAL`
- `EXPIRED`

## Migraciones

GORM ejecuta `AutoMigrate` al iniciar para:

- `subscription_plans`
- `plan_features`
- `subscriptions`

## Ejecutar

```bash
go mod tidy
go run .
```

Servicio por defecto en `http://localhost:8081`.

## Notas

- `user_id` proviene de IAM y se persiste sin FK cruzada.
- `stripe_subscription_id` se guarda como referencia externa cuando aplique.
- No existen tablas de `payments`, `invoices`, `transactions` o `billing` en este servicio.
