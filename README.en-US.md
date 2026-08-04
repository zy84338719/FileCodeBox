# FileCodeBox v2

> A high-performance file/text sharing platform built for self-hosting scenarios, using a front-end and back-end separated architecture.

---

## 📌 Project Overview

FileCodeBox is a lightweight sharing service implemented using Go + Vue 3, adopting a front-end and back-end separated architecture:

- **Backend**: Go + CloudWeGo Hertz + GORM
- **Frontend**: Vue 3 + TypeScript + Vite + Element Plus

### What can it do for you?

- 📁 **Drag-and-drop upload**, generate short links to share files or text snippets
- 🔐 **Admin backend**, centrally view, search, statistics, review, delete shared content
- 🪶 **Multiple storage backends**, switch between local/object storage solutions as needed
- ⚙️ **Custom configuration**, adjust rate limits, expiration policies, etc.

---

## 📸 Interface Preview

**Home (File / Text Sharing)**

![Home](https://raw.githubusercontent.com/zy84338719/FileCodeBox/main/docs/screenshots/01-home.png)

**Anonymous Retrieval**

![Anonymous Retrieval](https://raw.githubusercontent.com/zy84338719/FileCodeBox/main/docs/screenshots/08-retrieve.png)

**Admin Backend - Dashboard**

![Dashboard](https://raw.githubusercontent.com/zy84338719/FileCodeBox/main/docs/screenshots/04-admin-dashboard.png)

**Admin Backend - File Management**

![File Management](https://raw.githubusercontent.com/zy84338719/FileCodeBox/main/docs/screenshots/05-admin-files.png)

**Admin Backend - User Management / Storage Management**

![User Management](https://raw.githubusercontent.com/zy84338719/FileCodeBox/main/docs/screenshots/06-admin-users.png)

![Storage Management](https://raw.githubusercontent.com/zy84338719/FileCodeBox/main/docs/screenshots/07-admin-storage.png)

**Login Page (User / Admin)**

![User Login](https://raw.githubusercontent.com/zy84338719/FileCodeBox/main/docs/screenshots/02-user-login.png)

![Admin Login](https://raw.githubusercontent.com/zy84338719/FileCodeBox/main/docs/screenshots/03-admin-login.png)

---

## 🌟 Key Features

| Category | Feature Overview |
| --- | --- |
| Performance | Go native concurrency, chunked upload, resume upload, instant upload verification |
| Sharing Experience | Text/file dual channels, link validity period control, password and access count restrictions, **anonymous retrieval (vastsa UX)** |
| Admin Backend | Dashboard, file list, user management, storage panel, system configuration, **system notification announcements** |
| Security | JWT authentication (fail-fast in all environments + concurrency safe), API Key support, **rate limiting middleware (path-aware for login/upload/download)**, **CORS defaults to localhost only**, **security response headers** (HSTS/X-Frame-Options, etc.), **anonymous retrieval password bcrypt**, **upload size + type blacklist validation**, **metrics disabled by default + intranet binding** |
| Storage | Local disk, S3-compatible object storage, WebDAV, NFS, **OpenDAL-style abstraction** |
| Upload | Direct upload + **pre-signed upload** (init/complete/abort) |
| Observability | **Prometheus metrics (HTTP RED)**, **structured access logs (trace_id throughout the chain)**, **deep readiness probe (/readyz)** |
| Engineering | **Unified error code system**, **unified response envelope** (code/message/data/trace_id), **transaction guarantees for core write operations**, **full structured logging**, **120+ unit tests covering core business** |
| Deployment | **Single binary** (built-in frontend SPA), Docker, **production Compose (with Redis+Nginx)**, **K8s (probe/Ingress/HPA/ServiceMonitor)**, **environment variable configuration (12-factor)** |
| Frontend | Vue 3 + TypeScript, responsive layout, modern UI |

---

## 🧩 Project Structure

```
FileCodeBox/
├── backend/           # Go backend (Hertz v0.9.6 + Thrift IDL)
│   ├── cmd/server/    # Entry point + bootstrap
│   ├── internal/      # Internal packages
│   │   ├── app/       # Business logic (user/admin/share/anonymous/presign/notify/...)
│   │   ├── repo/      # Data access (db/redis)
│   │   ├── conf/      # Configuration
│   │   ├── pkg/
│   │   │   ├── errcode/   # Business error code ranges (1xxxx/2xxxx/...)
│   │   │   ├── resp/      # Unified response envelope (with trace_id)
│   │   │   ├── middleware # auth + ratelimit
│   │   │   └── ...
│   │   └── storage/
│   │       ├── storage.go     # Older 4-driver abstraction (local/s3/webdav/nfs)
│   │       └── opendal/      # OpenDAL-style abstraction
│   ├── gen/http/      # hz generated (thrift)
│   │   ├── handler/   # HTTP handlers (stub + business coverage)
│   │   ├── model/     # Models generated from Thrift
│   │   └── router/    # Route registration
│   ├── idl/           # Thrift IDL (fully migrated from proto)
│   ├── scripts/       # hz generation scripts + tool installation
│   └── configs/       # Configuration files
│
├── frontend/          # Vue 3 frontend
│   ├── src/
│   │   ├── views/     # Page components
│   │   ├── api/       # API calls
│   │   ├── stores/    # State management
│   │   └── router/    # Routing
│   └── dist/          # Build artifacts
│
├── Makefile           # Build scripts
├── docker-compose.yml # Docker Compose configuration
└── README.md
```

---

## 🚀 Quick Start

### 1. Environment Requirements

- **Go** 1.21+
- **Node.js** 18+
- **SQLite / MySQL / PostgreSQL** (defaults to SQLite)
- Docker 20+ (optional)

### 2. Local Development

```bash
# Install dependencies
make deps

# Development mode (front and back start simultaneously)
make dev

# Or start separately
make dev-backend   # Backend :12345
make dev-frontend  # Frontend :5173
```

### 3. Production Build

```bash
# Full build
make build

# Or build step by step
make build-frontend  # Build frontend
make build-backend   # Build backend
make copy-frontend   # Copy frontend to backend/static/
```

### 4. Production Deployment (4 ways, choose one)

#### Option A: Single Binary (Simplest, built-in frontend SPA)

```bash
make build                # Build frontend + backend (frontend embedded in static/)
cd backend && ./bin/server  # Single binary opens frontend + all APIs
```

#### Option B: Docker Single Container

```bash
docker run -d --name filecodebox \
  -p 12345:12345 \
  -v ./data:/app/data \
  -e FCB_PRODUCTION=1 \
  -e FCB_JWT_SECRET=$(openssl rand -hex 32) \
  ghcr.io/zy84338719/filecodebox:latest
```

#### Option C: Production Compose (Backend + Redis + Nginx full stack)

```bash
cp .env.example .env       # Fill in JWT_SECRET and other secrets
docker compose -f docker-compose.prod.yml up -d
```

#### Option D: Kubernetes (Production-grade, with probes/scaling/monitoring)

```bash
# 1. Edit deploy/k8s/base/secret.yaml to fill in real JWT_SECRET
# 2. Edit deploy/k8s/base/ingress.yaml to fill in domain name
kubectl apply -k deploy/k8s/base        # Or production overlay:
kubectl apply -k deploy/k8s/overlays/prod
```

> The service defaults to listening on `http://0.0.0.0:12345`, and the initial default administrator is `admin / admin123` (**change immediately in production**).

### 5. Configuration and Environment Variables

Configuration priority: **Environment Variables > Configuration File (yaml) > Default Values**. Sensitive configurations (keys/passwords) are recommended to be injected via environment variables, see [`docs/ENVIRONMENT_VARIABLES.md`](docs/ENVIRONMENT_VARIABLES.md).

```bash
# Common environment variables (both FCB_ prefixed full names and short names are acceptable)
FCB_SERVER_PORT=12345        # Service port
FCB_JWT_SECRET=xxx           # JWT key (⚠️ Required in all environments! If not set or set to default value, the service will fail to start and exit with error)
FCB_DATABASE_DRIVER=sqlite   # sqlite/mysql/postgres
FCB_REDIS_HOST=redis         # Redis address (anonymous retrieval function depends on Redis)
FCB_METRICS_ENABLED=false    # Prometheus metrics switch (disabled by default, when enabled, bind to intranet)
FCB_METRICS_ADDR=127.0.0.1:9090  # Metrics separate listening address (takes effect when enabled)
FCB_CORS_ALLOW_ORIGINS=https://your-domain.com  # CORS whitelist (if not configured, defaults to localhost only)
```

> ⚠️ **Security Tip**: `FCB_JWT_SECRET` must be set to a strong random value (≥32 characters) in all environments (including development), otherwise it will fail-fast on startup. Generation method: `openssl rand -hex 32`.

See the complete configuration template [`backend/configs/config.example.yaml`](backend/configs/config.example.yaml).

### 6. Observability

| Endpoint | Purpose |
| --- | --- |
| `GET 127.0.0.1:9090/metrics` | Prometheus metrics (HTTP RED, **disabled by default, when enabled, bind to intranet separate port, not exposed on main port**) |
| `GET /readyz` | Deep readiness check (includes DB ping, used for K8s readinessProbe) |
| `GET /live` | Liveness check (lightweight, used for livenessProbe) |
| `GET /health` | Health check |
| `GET /openapi.json` | OpenAPI 3.0 specification (Swagger UI) |

All logs are zap structured JSON, carrying `trace_id` (transparency of `X-Trace-Id` header), making it convenient for end-to-end troubleshooting.

---

## 📚 API Overview

### Public API

| Interface | Description |
| --- | --- |
| `POST /share/text/` | Share text |
| `POST /share/file/` | Share file |
| `GET /share/select/?code=...` | Get shared content |
| `GET /share/download` | Download file |
| `POST /user/register` | User registration |
| `POST /user/login` | User login |
| `GET /health` | Health check |

### Authenticated API (Requires JWT Token)

| Interface | Description |
| --- | --- |
| `GET /user/info` | User information |
| `POST /api/v1/user/refresh` | **Refresh token** (front-end 401 interceptor calls automatically) |
| `GET /user/files` | User file list |
| `GET /user/api-keys` | API Key list |
| `POST /user/api-keys` | Create API Key |

### Admin API (Requires Admin Token)

| Interface | Description |
| --- | --- |
| `POST /admin/login` | Admin login |
| `GET /admin/stats` | System statistics |
| `GET /admin/files` | File list |
| `GET /admin/users` | User list |
| `GET /admin/storage` | Storage information |

### Chunked Upload

| Interface | Description |
| --- | --- |
| `POST /chunk/upload/init/` | Initialize upload |
| `POST /chunk/upload/chunk/:id/:idx` | Upload chunk |
| `POST /chunk/upload/complete/:id` | Complete upload |
| `GET /chunk/upload/status/:id` | Upload status |
| `DELETE /chunk/upload/cancel/:id` | Cancel upload |

### Pre-signed Upload & Anonymous Retrieval

| Interface | Description |
| --- | --- |
| `POST /api/v1/presign/upload` | Apply for pre-signed upload (returns upload_id + token) |
| `PUT /api/v1/presign/upload-direct/:uploadID` | **Pre-signed direct upload** (with X-Upload-Token, write file to storage) |
| `POST /api/v1/presign/complete` | Complete upload (write share table) |
| `POST /anonymous/generate` | Anonymously generate 6-digit retrieval code (requires Redis) |
| `POST /anonymous/retrieve` | Anonymous retrieval (verify password + deduct count, final based on DB) |
| `GET /anonymous/search/:code` | Anonymous search for share information |

---

## 🔐 Authentication Methods

1. **JWT Token**: Obtained after user/admin login, placed in the `Authorization: Bearer <token>` header. When the token expires, the front-end automatically calls `/api/v1/user/refresh` to refresh.
2. **API Key**: Format `fcb_sk_xxx`, placed in the `X-API-Key` header or `api_key` query parameter.

---

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                      Browser (Vue 3 SPA)                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────┐ │
│  │ File Share │  │ Anonymous Retrieval │  │ Admin Backend │  │ User Center │ │
│  └─────┬────┘  └─────┬────┘  └─────┬────┘  └────┬────┘ │
└────────┼────────────┼────────────┼────────────┼────────┘
         │            │            │            │
         ▼            ▼            ▼            ▼
┌─────────────────────────────────────────────────────────┐
│              Hertz HTTP Server (:12345)                  │
│  ┌────────────────────────────────────────────────────┐ │
│  │ Middleware Chain: Recovery → RequestID → AccessLog →       │ │
│  │   Metrics → SecurityHeaders → CORS → Rate Limiting (Path-Aware)│ │
│  └────────────────────────────────────────────────────┘ │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐ │
│  │  app/share  │  │app/anonymous│  │   app/admin     │ │
│  │ (Share/Retrieve) │  │ (6-digit retrieval code) │  │  (Backend/Statistics)    │ │
│  └──────┬──────┘  └──────┬──────┘  └────────┬────────┘ │
│         │                │                  │          │
│  ┌──────▼────────────────▼──────────────────▼────────┐ │
│  │   DAO Layer (dao.FileCodeRepository / Notify / ...)  │ │
│  │   All pass through global db.GetDB(), transaction guarantee for write consistency          │ │
│  └──────┬──────────────────────────────┬─────────────┘ │
└─────────┼──────────────────────────────┼───────────────┘
          ▼                              ▼
   ┌────────────┐               ┌────────────────┐
   │  Database    │               │     Redis      │
   │ SQLite/     │               │ Anonymous retrieval code mapping │
   │ MySQL/PG    │               │ presign meta   │
   └────────────┘               └────────────────┘
          ▲
          │
   ┌──────┴──────┐   ┌──────────────┐
   │  Storage Abstraction   │   │ metrics separate │
   │ local/s3/   │   │ :9090(intranet)  │
   │ webdav/nfs  │   │ Disabled by default     │
   └─────────────┘   └──────────────┘
```

> 📊 Detailed test validation see [Local Test Report](docs/LOCAL-TEST-REPORT.md)

---

## 🛠️ Configuration Instructions

The configuration file is located at `backend/configs/config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 12345

database:
  driver: "sqlite"
  db_name: "./data/filecodebox.db"

storage:
  type: "local"
  path: "./data/uploads"

user:
  allow_user_registration: true
```

---

## 🧑‍💻 Development Guide

### Adding New API

1. Add a `.thrift` file in `backend/idl/http/`
2. Run `make thrift-gen IDL=idl/http/your_api.thrift` or `make thrift-gen-all` (used in CI)
3. Implement the handler in `backend/gen/http/handler/`
4. Add business logic in `backend/internal/app/`

### Modifying Frontend

1. Modify the code under `frontend/src/`
2. Run `make build-frontend` to build
3. Run `make copy-frontend` to copy to the backend

---

## 🚧 Known Limitations / Future Plans

| Item | Status | Description |
| --- | --- | --- |
| OpenDAL real binding | Pending Linux deployment | macOS/Windows compilation goes through stub; Linux production environment enables Apache OpenDAL Go binding, connecting to S3/OSS/WebDAV/HDFS, etc. |
| proto IDL | ✅ Fully migrated to thrift | proto → thrift migration completed, CI uses `make thrift-gen-all` |
| Frontend i18n | ✅ Completed | vue-i18n integrated (zh-CN / en-US) |
| Frontend dark mode | ✅ Completed | Dark mode + top bar toggle |
| Environment variable configuration | ✅ Completed | viper env binding (FCB_ prefix + short name), 12-factor compliant |
| Production observability | ✅ Completed | Prometheus metrics + structured access logs (trace_id) + deep readiness probe |
| Security hardening | ✅ Completed | Password bcrypt + rate limiting mounted + JWT fail-fast in all environments + CORS tightened + upload validation + metrics intranet isolation + frontend XSS disinfection |
| K8s production deployment | ✅ Completed | Probes/Ingress/HPA/ServiceMonitor/Secret/ConfigMap + production overlay |
| Business unit test coverage | ✅ Completed | 12 packages, 120+ cases, covering core business (including transactions/atomic deduction/password verification/anonymous retrieval DB pass-through) |
| Database version migration | Pending | Currently uses GORM AutoMigrate; enterprise-level can introduce golang-migrate for version baseline |
| Rate limiting persistence | Pending | Current ratelimit is memory-based; multi-instance needs Redis persistence to ensure consistency |
| OpenTelemetry tracing | Pending | Already reserved observability.tracing configuration bit, OTel middleware + OTLP exporter pending integration |
| Rate limiting stress test | Pending stress test platform | Rate limiting middleware has been implemented and covered by unit tests; QPS threshold calibration needs k6/wrk verification |

---

## 📄 License

MIT License © FileCodeBox Contributors
