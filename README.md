<div align="center">

# Recho

### A real-time chat application built with Go, React, PostgreSQL, and WebSockets.

Recho is a full-stack real-time messaging application designed around a Go backend and a React + TypeScript frontend. It supports authentication, direct and group conversations, realtime messaging, message editing and deletion, media storage, group invitations, and cursor-based message history.

<br />

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge\&logo=go\&logoColor=white)](https://go.dev/)
[![Echo](https://img.shields.io/badge/Echo-v4-3E863D?style=for-the-badge\&logo=go\&logoColor=white)](https://echo.labstack.com/)
[![React](https://img.shields.io/badge/React-19.2-61DAFB?style=for-the-badge\&logo=react\&logoColor=white)](https://react.dev/)
[![TypeScript](https://img.shields.io/badge/TypeScript-6.0-3178C6?style=for-the-badge\&logo=typescript\&logoColor=white)](https://www.typescriptlang.org/)
[![Vite](https://img.shields.io/badge/Vite-8.2-646CFF?style=for-the-badge\&logo=vite\&logoColor=white)](https://vite.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-336791?style=for-the-badge\&logo=postgresql\&logoColor=white)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge\&logo=docker\&logoColor=white)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)

<br />

</div>

---

## Features

* 🔐 **Authentication** — Registration and login with JWT-based authentication.
* 💬 **Direct Messaging** — Real-time one-to-one conversations with message editing, deletion, and cursor-based history.
* 👥 **Group Chats** — Real-time group conversations
* 🔗 **Group Invitations** — Shareable invite links with group preview before joining.
* 👤 **User Profiles** — Search users and edit personal profile information, including display name, bio, username, and avatar.
* 🏷️ **Group Profiles** — Edit group information such as name, bio, and avatar.
* ⚡ **Real-time Updates** — WebSocket based real-time messaging and conversation updates.
* 📎 **File Storage** — Support for local storage and S3-compatible object storage.
* 🌓 **Modern UI** — Responsive interface with light/dark themes and emoji support.

---

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

---

# Getting Started

## Prerequisites

Make sure the following are installed:

* Go 1.26+
* Node.js 20+
* npm
* Docker
* Docker Compose
* `swag` CLI *(optional — only required when regenerating Swagger documentation)*

---

## 1. Clone the Repository

```bash
git clone https://github.com/Hossein-Fazel/Recho.git
cd Recho
```

---

## 2. Configure Environment

Create your local environment file:

```bash
cp .env.example .env
```

Review the configuration before starting the application.

---

# Running Recho

There are two recommended development workflows.

### Option A

Run PostgreSQL, the Go backend, and the built React application through Docker Compose.

### Option B

Run PostgreSQL through Docker while running the Go backend and React frontend directly on your machine.

For active development, Option B provides frontend hot reload and faster backend iteration.

---

## Option A: Run Everything with Docker Compose

Set:

```env
SERVER_MODE=production
```

Then run:

```bash
docker compose up --build
```

Docker Compose will:

1. Build the application image.
2. Build the React frontend.
3. Compile the Go backend.
4. Start PostgreSQL 17.
5. Wait for PostgreSQL to become healthy.
6. Run database migrations.
7. Start the application.
8. Serve the API and compiled frontend together.

Open:

```text
http://localhost:8000
```

The exact port is controlled by `SERVER_PORT`.

### Run in background

```bash
docker compose up --build -d
```

### Stop containers

```bash
docker compose down
```

### Stop containers and remove database volume

```bash
docker compose down -v
```

> Docker mode serves the compiled frontend. It does not run the Vite development server, so frontend hot reload is not available.

---

## Option B: PostgreSQL in Docker, Backend & Frontend Locally

This workflow is recommended for active development.

### Start PostgreSQL

```bash
docker compose up postgres -d
```

### Start the backend

```bash
make run
```

The API will be available at:

```text
http://localhost:8000
```

Database migrations are applied automatically when the server starts.

In development mode, Swagger UI is available at:

```text
http://localhost:8000/{SERVER_SWAGGER_ROUTE}/index.html
```

---

### Start the frontend

In another terminal:

```bash
cd UI
npm install
npm run dev
```

The Vite development server runs at:

```text
http://localhost:5173
```

The Vite development configuration proxies:

```text
/api → backend
/ws  → backend
```

This keeps API and WebSocket requests effectively same-origin from the browser's perspective during development.

---

# Configuration

## Server & Database

| Variable               | Description                   | Default       |
| ---------------------- | ----------------------------- | ------------- |
| `POSTGRES_HOST`        | PostgreSQL host               | `localhost`   |
| `POSTGRES_PORT`        | PostgreSQL port               | `5432`        |
| `POSTGRES_USER`        | PostgreSQL username           | `user`        |
| `POSTGRES_PASSWORD`    | PostgreSQL password           | `pass`        |
| `POSTGRES_NAME`        | Database name                 | `DB_Name`     |
| `POSTGRES_SSLMODE`     | PostgreSQL SSL mode           | `disable`     |
| `SERVER_PORT`          | HTTP server port              | `8000`        |
| `SERVER_SWAGGER_ROUTE` | Swagger route prefix          | `swagger`     |
| `SERVER_MODE`          | `development` or `production` | `development` |
| `TOKEN_SECRET`         | JWT signing secret            | `mysecret`    |
| `TOKEN_ISSUER`         | JWT issuer                    | `Recho`       |
| `TOKEN_RT_TTL`         | Refresh-token TTL             | `720h`        |
| `TOKEN_AT_TTL`         | Access-token TTL              | `1h`          |
| `STORAGE_DRIVER`       | `local` or `s3`               | `local`       |

### Local Storage

| Variable                  | Description                 | Default                         |
| ------------------------- | --------------------------- | ------------------------------- |
| `STORAGE_LOCAL_ROOT_PATH` | Upload directory            | `./uploads`                     |
| `STORAGE_LOCAL_BASE_URL`  | Base URL for uploaded files | `http://localhost:8000/uploads` |

### S3 / S3-Compatible Storage

| Variable                       | Description                      | Default      |
| ------------------------------ | -------------------------------- | ------------ |
| `STORAGE_S3_ENDPOINT`          | Custom S3 endpoint               | Empty        |
| `STORAGE_S3_REGION`            | Bucket region                    | `us-east-1`  |
| `STORAGE_S3_BUCKET`            | Bucket name                      | `recho`      |
| `STORAGE_S3_ACCESS_KEY_ID`     | Access key                       | `minioadmin` |
| `STORAGE_S3_SECRET_ACCESS_KEY` | Secret key                       | `minioadmin` |
| `STORAGE_S3_USE_PATH_STYLE`    | Use path-style addressing        | `true`       |
| `STORAGE_S3_DISABLE_ACL`       | Disable public-read ACL handling | `false`      |
| `STORAGE_S3_PUBLIC_BASE_URL`   | Public/CDN base URL              | Empty        |
| `STORAGE_S3_PRESIGN_EXPIRY`    | Presigned URL lifetime           | `15m`        |

> **Security:** The values shown above are development defaults. Replace secrets and credentials before using Recho in a non-local environment.

---

# Running with MinIO

Recho includes a Docker Compose S3 profile for local development with MinIO.

Set:

```env
STORAGE_DRIVER=s3
```

Then run:

```bash
docker compose --profile s3 up --build
```

This starts:

* Recho
* PostgreSQL
* MinIO
* MinIO initialization

MinIO's web console is available at:

```text
http://localhost:9001
```

Inside the Docker network, the application should use:

```env
STORAGE_S3_ENDPOINT=http://minio:9000
```

For browser-accessible generated URLs, configure the public base URL using the host-accessible MinIO endpoint.

---

# Production Build

The application can also be built without Docker.

First build the frontend:

```bash
cd UI
npm install
npm run build
cd ..
```

Then build/run the Go application:

```bash
make build
```

The resulting application serves:

```text
/api/*
```

and the compiled:

```text
UI/dist
```

from the same Go process.

This allows Recho's backend and frontend to be deployed together rather than requiring separate application servers.

---

# API Documentation

Recho exposes its HTTP API through Swagger/OpenAPI documentation in development mode.

With the default configuration:

```text
http://localhost:8000/swagger/index.html
```

Swagger documentation is generated from Go annotations.

If you modify API annotations and need to regenerate the documentation, install the `swag` CLI and regenerate the Swagger files according to the project's development workflow.

---

# Contributing

Contributions are welcome.

For significant changes, please open an issue first to discuss the proposed change before submitting a pull request.

When contributing:

1. Create a focused branch for your change.
2. Keep changes scoped to the relevant feature or fix.
3. Follow the existing project structure and conventions.
4. Test your changes locally.
5. Update documentation when behavior or configuration changes.

---

# License

Recho is licensed under the [MIT License](LICENSE).
