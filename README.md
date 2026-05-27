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
- `STRIPE_PRICE_FREE`
- `STRIPE_PRICE_PLUS`
- `STRIPE_PRICE_PRO`
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

### Webhooks

- `POST /api/v1/webhooks/stripe`

## Estados de suscripcion

- `ACTIVE`
- `INACTIVE`
- `CANCELLED`
- `PENDING_RENEWAL`
- `EXPIRED`

## Migraciones

El servicio valida/crea las tablas requeridas al iniciar:

- `subscription_plans`
- `plan_features`
- `subscriptions`

## Ejecutar

```bash
go mod tidy
go run .
```

Servicio por defecto en `http://localhost:8081`.

## Docker

Construir la imagen:

```bash
docker build -t microservice-subscriptions-service .
```

Ejecutar localmente usando tu `.env`:

```bash
docker run --env-file .env -p 8082:8082 microservice-subscriptions-service
```

Si no defines `SERVER_PORT`, el contenedor usa `8081`:

```bash
docker run --env-file .env -p 8081:8081 microservice-subscriptions-service
```

## Deploy en Render

### Opcion recomendada: sin Docker

Este repositorio incluye `render.yaml` para desplegar como servicio Go nativo.

1. Sube el repo a GitHub.
2. En Render, crea un nuevo Blueprint o Web Service desde el repo.
3. Si usas Blueprint, Render leera `render.yaml`.
4. Agrega las variables marcadas como secretas:
   - `DATABASE_URL`
   - `STRIPE_SECRET_KEY`
   - `STRIPE_PUBLISHABLE_KEY`
   - `STRIPE_WEBHOOK_SECRET`
   - `STRIPE_PRICE_FREE`
   - `STRIPE_PRICE_PLUS`
   - `STRIPE_PRICE_PRO`
   - `KAFKA_BROKERS`
5. Build command: `go build -tags netgo -ldflags "-s -w" -o app .`
6. Start command: `./app`

Render inyecta `PORT` automaticamente y la app lo usa antes que `SERVER_PORT`.
El health check esta disponible en `GET /health`.

### Opcion con Docker

Tambien puedes crear el Web Service seleccionando runtime Docker. Render usara el
`Dockerfile` del repo y las mismas variables de entorno. No subas `.env` a Git ni
lo copies dentro de la imagen.

## Keep-Alive (Render/Free plans)

Puedes usar un pinger externo para mantener vivo el servicio:

- Windows PowerShell:
```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\keepalive.ps1 -Url "https://tu-servicio.onrender.com/api/v1/subscription-plans" -IntervalSeconds 600
```

- Linux/macOS:
```bash
bash ./scripts/keepalive.sh "https://tu-servicio.onrender.com/api/v1/subscription-plans" 600
```

## Notas

- `user_id` proviene de IAM y se persiste sin FK cruzada.
- `stripe_subscription_id` se guarda como referencia externa cuando aplique.
- Para crear/cambiar suscripcion en Stripe se espera `STRIPE_PRICE_ID` en `plan_features.feature_code`.
- Al iniciar, el servicio asegura planes base `Free`, `Plus`, `Pro` y sincroniza su `STRIPE_PRICE_ID` desde `STRIPE_PRICE_FREE`, `STRIPE_PRICE_PLUS`, `STRIPE_PRICE_PRO`.
- `POST /api/v1/subscriptions` acepta `stripe_customer_id` para crear suscripcion real en Stripe.
- No existen tablas de `payments`, `invoices`, `transactions` o `billing` en este servicio.
