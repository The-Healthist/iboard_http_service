# Copilot Instructions for iboard_http_service

These notes orient an AI coding assistant to the key patterns, workflows, and conventions in this Go-based HTTP service.

## 1. Architecture Overview

- **Entry point**: `cmd/server/main.go` loads configuration via `pkg/utils/config.go`, initializes router (`internal/app/router/router.go`), middleware and starts HTTP server.
- **Application layer**: `internal/app/controller/*` implements HTTP handlers. Controllers call domain services.
- **Domain layer**: `internal/domain/models` defines GORM models; `internal/domain/services` contains business logic and data-access orchestration.
- **Infrastructure**: `infrastructure/database/database.go` configures GORM MySQL; `infrastructure/redis/redis.go` sets up Redis client.
- **Utilities**: `pkg/utils/response` standardizes JSON responses; `pkg/validator` wraps go-playground/validator for request payload validation.

## 2. Developer Workflows

- **Configuration**: Environment variables in `.env` (loaded via `godotenv` in `cmd/server`). Key vars: `DB_HOST`, `DB_PORT`, `DB_USER/PASSWORD`, `REDIS_HOST/PORT`, `CALLBACK_URL_*`, `ACCESS_KEY_*`.
- **Docker Compose**:
  ```bash
  # Build with embedded version (default 1.1.0)
  docker compose build --no-cache backend
  # Launch all services (backend, mysql, redis)
  docker compose up -d
  # Tail logs
  docker compose logs -f backend
  ```
- **Local Go tests**:
  ```bash
  go test ./internal/domain/... ./pkg/utils/... -timeout 30s
  ```
- **Database migrations**: SQL scripts in `migrations/`; helper scripts in `scripts/migrate/`.

## 3. Project Conventions

- **Controllers**: under `internal/app/controller`, one file per resource. Use Gin binding tags (`json`, `form`) and call services in `internal/domain/services`.
- **JSON fields**: multi-value columns stored as `gorm.io/datatypes.JSON`. Look at `models/device.go` for examples.
- **Error handling**: return standardized response structs (see `pkg/utils/response.Response`), use HTTP status codes in controllers.
- **Middleware**: CORS (`internal/app/middleware/cors.go`), JWT auth (`internal/app/middleware/jwt.go`). Always apply in `router/router.go`.
- **Logging**: uses `pkg/log/gin_logger.go` to integrate with Gin.

## 4. Integration Points & External APIs

- **MySQL**: GORM auto-migrations disabled by default; database schema managed by raw SQL in `migrations/`. DSN uses `mysql_native_password`.
- **Redis**: go-redis v8, client instantiation in `infrastructure/redis/redis.go`, DB index from `REDIS_DB`.
- **Swagger**: docs in `docs/docs_swagger`; swagger YAML at `swagger.yaml` and router registration in `docs/docs_swagger/docs.go`.
- **OSS Storage**: Aliyun keys `ACCESS_KEY_ID/SECRET`, used in `utils/email.go` and file upload controller.

## 5. Code Patterns & Examples

- **Routing**: all routes declared in `internal/app/router/router.go` with grouped paths and middleware.
- **Service call**:
  ```go
  func GetBuilding(c *gin.Context) {
    id := c.Param("id")
    result, err := services.BuildingService.FindByID(id)
    c.JSON(http.StatusOK, response.New(result, err))
  }
  ```
- **Validation**: in controller, call `validator.ValidateStruct(&payload)` before service.

---

Please review and let me know if any areas are unclear or need deeper detail!
