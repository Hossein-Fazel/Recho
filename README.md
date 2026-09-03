# Recho

**Recho** is a realtime chat application built with a Go backend and a React (TypeScript) frontend. It provides user authentication, direct conversations, and realtime message delivery over WebSockets.

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white) ![Echo](https://img.shields.io/badge/Echo-000000?style=for-the-badge&logo=go&logoColor=white) ![React](https://img.shields.io/badge/React-61DAFB?style=for-the-badge&logo=react&logoColor=black) ![TypeScript](https://img.shields.io/badge/TypeScript-3178C6?style=for-the-badge&logo=typescript&logoColor=white) ![Vite](https://img.shields.io/badge/Vite-646CFF?style=for-the-badge&logo=vite&logoColor=white) ![PostgreSQL](https://img.shields.io/badge/PostgreSQL-4169E1?style=for-the-badge&logo=postgresql&logoColor=white) ![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)

## Features

- **Authentication** — registration, login, JWT access tokens, and rotating refresh tokens delivered via HTTP-only cookies
- **Direct conversations** — start or resume a one-on-one conversation with another user
- **Realtime messaging** — WebSocket-based message delivery with a hub/client architecture
- **User search** — look up other users to start a conversation
- **Message history** — cursor-based pagination for fetching past messages
- **API documentation** — Swagger/OpenAPI docs generated from code annotations
- **Database migrations** — versioned Postgres schema managed with `golang-migrate`

## Architecture

Recho follows a clean/hexagonal architecture on the backend:

```
internal/
├── model/         # Core domain entities
├── application/   # Business logic and service interfaces
├── adapter/       # Interface implementations (e.g. Postgres repositories)
├── infra/         # Infrastructure concerns (DB pool, migrations, token signing)
├── delivery/
│   ├── web/         # HTTP layer (Echo handlers, DTOs, middleware, error handling)
│   └── websocket/   # WebSocket hub, client, and message delivery
├── config/        # Environment-based configuration loading
└── apperr/        # Application-level error types
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
