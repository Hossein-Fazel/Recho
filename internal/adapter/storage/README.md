
# Recho Storage

Recho uses a storage abstraction to handle uploaded files such as avatars, images, voice messages, videos, and other files.

The application does not depend directly on a specific storage provider. Instead, it uses a common storage interface and selects the implementation through configuration.

Currently supported drivers:

* **Local** — stores files on the server filesystem.
* **S3** — stores files in Amazon S3 or any S3-compatible object storage such as MinIO.

The storage driver is selected using:

```env
STORAGE_DRIVER=local
```

or:

```env
STORAGE_DRIVER=s3
```

---

## Configuration

Storage configuration is done through environment variables.

All available variables are listed in `.env.example`.

### General

| Variable           | Description            | Example            |
| ------------------ | ---------------------- | ------------------ |
| `STORAGE_DRIVER` | Storage backend to use | `local` / `s3` |

---

## Local Storage

The local driver stores uploaded files on the application's filesystem.

### Environment Variables

```env
STORAGE_DRIVER=local

STORAGE_LOCAL_ROOT_PATH=./uploads
STORAGE_LOCAL_BASE_URL=http://localhost:8000/uploads
```

| Variable                    | Description                               |
| --------------------------- | ----------------------------------------- |
| `STORAGE_LOCAL_ROOT_PATH` | Directory where uploaded files are stored |
| `STORAGE_LOCAL_BASE_URL`  | Base URL used when generating file URLs   |

For example:

```text
STORAGE_LOCAL_ROOT_PATH=./uploads
```

means files will be stored under:

```text
uploads/
```

If:

```text
STORAGE_LOCAL_BASE_URL=http://localhost:8000/uploads
```

and an object has the key:

```text
avatars/users/123/avatar.png
```

its URL will be:

```text
http://localhost:8000/uploads/avatars/users/123/avatar.png
```

### Docker

When using Docker Compose, the uploads directory is persisted using a Docker volume.

This prevents uploaded files from being lost when the application container is recreated.

---

# S3 Storage

The S3 driver supports both Amazon S3 and S3-compatible object storage.

Examples:

* Amazon S3
* MinIO
* Cloudflare R2
* DigitalOcean Spaces

### Environment Variables

```env
STORAGE_DRIVER=s3

STORAGE_S3_ENDPOINT=
STORAGE_S3_REGION=us-east-1
STORAGE_S3_BUCKET=recho
STORAGE_S3_ACCESS_KEY_ID=
STORAGE_S3_SECRET_ACCESS_KEY=
STORAGE_S3_USE_PATH_STYLE=true
STORAGE_S3_DISABLE_ACL=false
STORAGE_S3_PUBLIC_BASE_URL=
STORAGE_S3_PRESIGN_EXPIRY=15m
```

| Variable                         | Description                                              |
| -------------------------------- | -------------------------------------------------------- |
| `STORAGE_S3_ENDPOINT`          | Custom S3-compatible endpoint. Leave empty for Amazon S3 |
| `STORAGE_S3_REGION`            | S3 region                                                |
| `STORAGE_S3_BUCKET`            | Bucket name                                              |
| `STORAGE_S3_ACCESS_KEY_ID`     | S3 access key                                            |
| `STORAGE_S3_SECRET_ACCESS_KEY` | S3 secret key                                            |
| `STORAGE_S3_USE_PATH_STYLE`    | Use path-style S3 requests                               |
| `STORAGE_S3_DISABLE_ACL`       | Disable ACL handling                                     |
| `STORAGE_S3_PUBLIC_BASE_URL`   | Public URL/CDN base URL                                  |
| `STORAGE_S3_PRESIGN_EXPIRY`    | Lifetime of generated presigned URLs                     |

---

## Amazon S3

For Amazon S3, the endpoint can be left empty:

```env
STORAGE_DRIVER=s3

STORAGE_S3_ENDPOINT=
STORAGE_S3_REGION=us-east-1
STORAGE_S3_BUCKET=recho
STORAGE_S3_ACCESS_KEY_ID=your-access-key
STORAGE_S3_SECRET_ACCESS_KEY=your-secret-key
STORAGE_S3_USE_PATH_STYLE=false
STORAGE_S3_DISABLE_ACL=true
STORAGE_S3_PUBLIC_BASE_URL=
STORAGE_S3_PRESIGN_EXPIRY=15m
```

For production, use the credential mechanism recommended by AWS rather than committing credentials to `.env`.

---

# MinIO

MinIO can be used locally as an S3-compatible storage backend.

Recho's Docker Compose configuration includes an S3 profile for running MinIO.

Start it with:

```bash
docker compose --profile s3 up --build
```

Use:

```env
STORAGE_DRIVER=s3

STORAGE_S3_ENDPOINT=http://minio:9000
STORAGE_S3_REGION=us-east-1
STORAGE_S3_BUCKET=recho
STORAGE_S3_ACCESS_KEY_ID=minioadmin
STORAGE_S3_SECRET_ACCESS_KEY=minioadmin
STORAGE_S3_USE_PATH_STYLE=true
STORAGE_S3_DISABLE_ACL=false
STORAGE_S3_PUBLIC_BASE_URL=http://localhost:9000/recho
STORAGE_S3_PRESIGN_EXPIRY=15m
```

### Important

There are two different MinIO addresses:

```text
http://minio:9000
```

is used by the Recho container to communicate with MinIO through the Docker network.

```text
http://localhost:9000
```

is used by the browser/host to access MinIO.

Therefore, when running Recho and MinIO through Docker, `STORAGE_S3_ENDPOINT` should normally use:

```env
STORAGE_S3_ENDPOINT=http://minio:9000
```

while the public base URL can use:

```env
STORAGE_S3_PUBLIC_BASE_URL=http://localhost:9000/recho
```

The MinIO console is available at:

```text
http://localhost:9001
```

> The default MinIO credentials are intended for local development only.

---

# Public and Private Files

Recho distinguishes between public and private files.

### Public

Avatars are public and can use a normal storage URL.

### Private

Message media such as images, voice messages, videos, and files are private.

For S3 storage, private files can be accessed using temporary presigned URLs.

The lifetime of these URLs is configured with:

```env
STORAGE_S3_PRESIGN_EXPIRY=15m
```

For example:

```env
STORAGE_S3_PRESIGN_EXPIRY=1h
```

creates URLs that are valid for approximately one hour.

---

# Public Base URL

`STORAGE_S3_PUBLIC_BASE_URL` can be used when files should be accessed through a custom domain or CDN.

For example:

```env
STORAGE_S3_PUBLIC_BASE_URL=https://cdn.example.com
```

An object with the key:

```text
avatars/users/123/avatar.png
```

will use:

```text
https://cdn.example.com/avatars/users/123/avatar.png
```

This is useful when the actual S3 endpoint should not be exposed to clients.

---

# Upload Limits

The application validates uploaded files before storing them.

Current limits are:

| Type   | Maximum |
| ------ | ------: |
| Avatar |    5 MB |
| Image  |   25 MB |
| Voice  |  100 MB |
| Video  |    5 GB |
| File   |   50 GB |

These limits are defined in:

```text
internal/application/storage_policy.go
```

---

# Storage Structure

Files are stored using generated object keys rather than their original filenames.

For example:

```text
avatars/users/<user-id>/<uuid>.png
```

or:

```text
conversations/<conversation-id>/image/2026/09/<uuid>.webp
```

The database stores the **object key**, not the complete storage URL.

This allows the storage provider or public URL to be changed without having to update every stored file reference.

---

# Switching Drivers

The storage backend can be changed through `STORAGE_DRIVER`.

For local storage:

```env
STORAGE_DRIVER=local
```

For S3:

```env
STORAGE_DRIVER=s3
```

No application-level upload code needs to change when switching between the supported drivers.

---

## Related Files

```text
internal/
├── adapter/
│   └── storage/
│       ├── storage.go
│       ├── local.go
│       ├── s3.go
│       └── README.md
│
├── application/
│   ├── storage.go
│   └── storage_policy.go
│
└── model/
    └── media.go
```

* `storage.go` — storage driver configuration and initialization
* `local.go` — local filesystem implementation
* `s3.go` — S3 implementation
* `application/storage.go` — application-level upload and storage operations
* `application/storage_policy.go` — upload limits and media policies
* `model/media.go` — media-related types and helpers

For the main project documentation, see the [root README](../../../README.md).
