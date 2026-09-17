<div align="center">

  # Recho

**Recho** is a realtime chat application built with a Go backend and a React (TypeScript) frontend. It provides user authentication, direct and group conversations, and realtime message delivery over WebSockets.


[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge\&logo=go\&logoColor=white)](https://go.dev)
[![Echo](https://img.shields.io/badge/Echo-v4-3E863D?style=for-the-badge\&logo=go\&logoColor=white)](https://echo.labstack.com/)
[![React](https://img.shields.io/badge/React-19.2-61DAFB?style=for-the-badge\&logo=react\&logoColor=white)](https://react.dev/)
[![TypeScript](https://img.shields.io/badge/TypeScript-6.0-3178C6?style=for-the-badge\&logo=typescript\&logoColor=white)](https://www.typescriptlang.org/)
[![Vite](https://img.shields.io/badge/Vite-8.2-646CFF?style=for-the-badge\&logo=vite\&logoColor=white)](https://vite.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-336791?style=for-the-badge\&logo=postgresql\&logoColor=white)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?style=for-the-badge\&logo=docker\&logoColor=white)](https://www.docker.com/)

</div>

## Features

Recho is a full-featured realtime chat application. Below is an overview of what it offers.

### Authentication & Security

Users can register and log in with JWT-based authentication. Sessions are backed by short-lived access tokens that authorize API requests, paired with rotating refresh tokens delivered through secure, HTTP-only cookies so they stay out of reach of client-side scripts.

### Messaging

Recho supports one-on-one direct conversations that can be started or resumed at any time. Messages are delivered in realtime over WebSockets using a hub/client architecture, so every participant sees updates immediately. You can edit and delete your own messages, with the changes synchronized to everyone in realtime. Message history is loaded efficiently with cursor-based pagination, and optimistic messaging makes your own messages appear instantly while showing their delivery status.

### Groups

Group chats let you create and participate in realtime conversations with multiple people. Group members can be browsed directly from the conversation info panel, and each member holds a role — owner, admin, or member — shown with a visual badge. Every group has a unique invite code and a shareable `/join/<code>` link, and anyone can preview the group information before joining with a single action. Group conversations use Telegram-style sender grouping with avatars and display names.

### Users & Profiles

You can search for users and quickly start a conversation with them. User profiles, including bios and handles, can be viewed straight from the conversation, and detailed information is available for both direct and group conversations.

### User Experience

The frontend is a clean, responsive messaging interface with built-in dark and light themes, an emoji picker for adding emojis to messages, and visual delivery status feedback for optimistic messages.

### Backend & Infrastructure

In production, the Go server serves the compiled frontend as static assets from a single binary. The API is documented with Swagger/OpenAPI generated from code annotations, the PostgreSQL schema is version-controlled through `golang-migrate` migrations, and realtime connections are managed by a hub/client WebSocket architecture.


## Architecture

Recho follows a clean architecture on the backend:

```
internal/
├── model/           # Core domain entities
├── application/     # Business logic and service interfaces
├── adapter/         # Interface implementations (e.g. Postgres repositories)
├── infra/           # Infrastructure concerns (DB pool, migrations, token signing)
├── delivery/
│   ├── web/         # HTTP layer (Echo handlers, DTOs, middleware, error handling)
│   └── websocket/   # WebSocket hub, client, and message delivery
├── config/          # Environment-based configuration loading
└── apperr/          # Application-level error types
```

The compiled frontend (`UI/dist`) is served as static assets directly by the Echo server, so the API and UI can be deployed as a single binary.

## Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) 1.26+
- [Node.js](https://nodejs.org/) 20+ and npm
- [Docker](https://www.docker.com/) and Docker Compose (for PostgreSQL)
- [swag CLI](https://github.com/swaggo/swag) (optional, only needed to regenerate Swagger docs)

### 1. Clone the repository

```bash
git clone https://github.com/Hossein-Fazel/Recho.git
cd Recho
```

### 2. Configure environment variables

Copy the example environment file and adjust the values as needed:

```bash
cp .env.example .env
```

| Variable                 | Description                                                                                                    | Default         |
| ------------------------ | -------------------------------------------------------------------------------------------------------------- | --------------- |
| `POSTGRES_HOST`        | PostgreSQL host                                                                                                | `localhost`   |
| `POSTGRES_PORT`        | PostgreSQL port                                                                                                | `5432`        |
| `POSTGRES_USER`        | PostgreSQL username                                                                                            | `user`        |
| `POSTGRES_PASSWORD`    | PostgreSQL password                                                                                            | `pass`        |
| `POSTGRES_NAME`        | PostgreSQL database name                                                                                       | `DB_Name`     |
| `POSTGRES_SSLMODE`     | PostgreSQL SSL mode                                                                                            | `disable`     |
| `SERVER_PORT`          | Port the HTTP server listens on                                                                                | `8000`        |
| `SERVER_SWAGGER_ROUTE` | Route prefix for Swagger UI (dev mode only)                                                                    | `swagger`     |
| `SERVER_MODE`          | `development` (Swagger UI on, frontend not served) or `production` (Swagger UI off, built frontend served) | `development` |
| `TOKEN_SECRET`         | Secret used to sign JWTs                                                                                       | `mysecret`    |
| `TOKEN_ISSUER`         | JWT issuer claim                                                                                               | `Recho`       |
| `TOKEN_RT_TTL`         | Refresh token time-to-live                                                                                     | `720h`        |
| `TOKEN_AT_TTL`         | Access token time-to-live                                                                                      | `1h`          |
| `STORAGE_DRIVER`       | Storage backend to use: `local` or `s3`                                                                        | `local`       |

#### Storage

Recho stores uploaded media (avatars and message attachments) through a pluggable storage layer selected by `STORAGE_DRIVER`. With `local` (the default), files are written to disk and served by the Go server. With `s3`, files are uploaded to AWS S3 or any S3-compatible provider such as MinIO, Cloudflare R2, or DigitalOcean Spaces.

Local storage (`STORAGE_DRIVER=local`):

| Variable                     | Description                                                       | Default                              |
| ---------------------------- | ----------------------------------------------------------------- | ------------------------------------ |
| `STORAGE_LOCAL_ROOT_PATH`    | Directory where uploaded files are written on disk                | `./uploads`                          |
| `STORAGE_LOCAL_BASE_URL`     | Public base URL used to build file URLs                           | `http://localhost:8000/uploads`      |

S3 / S3-compatible storage (`STORAGE_DRIVER=s3`):

| Variable                       | Description                                                                                              | Default                     |
| ------------------------------ | -------------------------------------------------------------------------------------------------------- | --------------------------- |
| `STORAGE_S3_ENDPOINT`          | Custom S3 endpoint. Leave empty for AWS S3; set it for MinIO/R2/Spaces (e.g. `http://localhost:9000`)    | _(empty)_                   |
| `STORAGE_S3_REGION`            | Bucket region                                                                                            | `us-east-1`                 |
| `STORAGE_S3_BUCKET`            | Bucket name (required when using `s3`)                                                                   | `recho`                     |
| `STORAGE_S3_ACCESS_KEY_ID`     | Access key ID. Optional — falls back to the default AWS credential chain when unset                      | `minioadmin`                |
| `STORAGE_S3_SECRET_ACCESS_KEY` | Secret access key. Optional — falls back to the default AWS credential chain when unset                  | `minioadmin`                |
| `STORAGE_S3_USE_PATH_STYLE`    | Use path-style bucket addressing. Required by most S3-compatible providers                               | `true`                      |
| `STORAGE_S3_DISABLE_ACL`       | Skip setting public-read ACLs. Set `true` for buckets with ACLs disabled (new AWS S3 buckets)            | `false`                     |
| `STORAGE_S3_PUBLIC_BASE_URL`   | Optional CDN or custom domain used when building public file URLs                                        | _(empty)_                   |
| `STORAGE_S3_PRESIGN_EXPIRY`    | Time-to-live for generated presigned URLs                                                                | `15m`                       |

> **Note:** Change `TOKEN_SECRET` and the database credentials before deploying anywhere beyond your local machine.

You now have two ways to run Recho. **Option A** runs everything (Postgres + the Go server + the built frontend) with a single Docker Compose command — the fastest way to get a production-like build running. **Option B** runs Postgres in Docker but runs the backend and frontend directly on your machine with the Go and Node CLIs — better for active development, since it gives you live reload on the frontend and fast `go run` iteration on the backend.

### Option A: Everything with Docker Compose

```bash
docker compose up --build
```

This single command:

- Builds the app image from the [`Dockerfile`](Dockerfile) (compiling the Go binary and building the React frontend in a multi-stage build)
- Starts a PostgreSQL 17 container using the credentials from your `.env` file
- Starts the `app` container once Postgres reports healthy, running database migrations automatically on boot
- Serves the API and the built frontend together at `http://localhost:$SERVER_PORT` (default `http://localhost:8000`)

> **Important:** the server only serves the built frontend (and disables Swagger UI) when `SERVER_MODE=production`. The default `.env.example` ships with `SERVER_MODE=development`, which is meant for Option B below. Set `SERVER_MODE=production` in your `.env` before running Option A, or the app container will serve the API only, with no frontend.

Add `-d` to run in the background. To stop everything:

```bash
docker compose down
```

To stop and also delete the Postgres data volume:

```bash
docker compose down -v
```

> Note: this option serves the **built** frontend (static files), not the Vite dev server, so there's no frontend hot-reload here. Use Option B for that.

If you want to try the S3 storage backend without an external provider, Docker Compose ships a bundled MinIO instance behind the `s3` profile. Set `STORAGE_DRIVER=s3` in your `.env` and start it alongside the app:

```bash
docker compose --profile s3 up --build
```

This starts MinIO (with a console at `http://localhost:9001`) and a one-shot `minio-init` container that creates the bucket and makes it publicly readable. When running with Compose, point `STORAGE_S3_ENDPOINT` at the Compose service (`http://minio:9000`) so the app container can reach MinIO, and set `STORAGE_S3_PUBLIC_BASE_URL` to a browser-reachable address such as `http://localhost:9000/$STORAGE_S3_BUCKET` for generated file URLs. With the default `local` driver, uploads are instead written to the `uploads_data` volume mounted at `/app/uploads`.

### Option B: PostgreSQL in Docker, everything else via CLI

Start only the database container:

```bash
docker compose up postgres -d
```

This starts a PostgreSQL 17 container on the port from `POSTGRES_PORT` in your `.env` (default `5432`), with the `.env` values already set up to point at it via `POSTGRES_HOST=localhost`.

Run the backend:

```bash
make run
```

The API server starts on `http://localhost:$SERVER_PORT` (default `8000`) and applies database migrations automatically on startup. In development mode, Swagger UI is available at `/{SERVER_SWAGGER_ROUTE}/index.html`.

Run the frontend, in a separate terminal:

```bash
cd UI
npm install
npm run dev
```

The Vite dev server runs on `http://localhost:5173` with hot reload. It proxies `/api` and `/ws` requests to the backend at `http://localhost:8000` (see `UI/vite.config.ts`), so the browser sees everything as same-origin and no CORS configuration is needed in development.

#### Building for production (CLI)

If you want a production-style build without Docker, build the frontend and let the Go server serve it as static files:

```bash
cd UI
npm run build
cd ..
make build
```

`make build` compiles the Go binary and runs it, serving both the API (`/api/*`) and the built frontend (`UI/dist`) from a single process.

## Contributing

Contributions are welcome. Please open an issue to discuss significant changes before submitting a pull request.

## License

This project is licensed under the [MIT License](LICENSE).
