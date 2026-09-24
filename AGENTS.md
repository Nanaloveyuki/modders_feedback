# AGENTS.md

## Project

This repository is a self-hosted feedback board for the RimWorld 1.6 mod `nanaloveyuki.ratkin.hungerandhavoc` (Ratkin: Hunger and Havoc).

The product has three feedback categories:

- `bug`: error reports
- `feature`: feature requests
- `question`: general questions

Visitors can browse, search, filter, sort, and inspect feedback. Anyone can register with a username, password, and optional email or QQ number. Registered members can create feedback and edit their own content. They can also switch their own feedback between `open` and `withdrawn`. The administrator account seeded from deployment secrets can update any status, edit or delete feedback, and change mods and displayed site versions and icons.

## Architecture

```text
Browser
  -> Go HTTP server (:8080)
       -> /api/*       JSON API
       -> /*           built React static files
       -> SQLite       /data/feedback.db

React/Vite: web/src/main.tsx, web/src/App.tsx, web/src/api, web/src/components, web/src/style.css
Go service: server/cmd/feedback
  internal/config, internal/domain, internal/store, internal/auth, internal/httpx, internal/api
Container build: Dockerfile
Deployment: docker-compose.yml
```

The Go service owns authentication, validation, SQLite access, API routing, static-file serving, and SPA fallback. The frontend must treat the API as the source of truth; do not add a second client-side persistence mechanism.

## Repository Rules

- Preserve the existing Go + React + SQLite + Docker architecture unless the task explicitly changes it.
- Keep the interface adapted to RimWorld mod feedback. Bug reports should retain fields for game version, mod version, mod list/load order, reproduction details, and optional save link.
- Keep the three category values exactly `bug`, `feature`, and `question`; they are enforced by the SQLite check constraint and API validation.
- Keep status values exactly `open`, `in_progress`, `resolved`, `closed`, and `withdrawn`.
- Public registration accepts a username, password, and optional email or QQ number. Do not store plaintext passwords, localStorage authentication tokens, or credentials in source control.
- Passwords are stored only as bcrypt hashes. Sessions are HttpOnly, SameSite=Lax cookies containing signed JWTs.
- Maintain request size limits, unknown-field rejection, title/body length validation, category validation, and status validation when changing API code.
- Keep the container non-root, read-only, capability-dropped, and backed by the named `feedback-data` volume.
- Do not use `docker compose down -v` during normal development. It deletes the SQLite database and all feedback.
- Do not modify or delete unrelated user files, data, or Docker volumes.
- Avoid adding dependencies unless the existing standard library or current dependency set cannot solve the requirement.

## Local Frontend Development

Requirements: Node.js and npm.

```sh
cd web
npm ci
npm run build
npm run dev
```

Vite serves the frontend and proxies `/api` to `http://localhost:8080`.

The API must run separately for development:

```sh
cd server
DATA_DIR=../data \
ADMIN_USERNAME=admin \
ADMIN_PASSWORD='admin123456_' \
JWT_SECRET='local-random-secret-at-least-32-characters' \
go run ./cmd/feedback
```

Run Vite from `web/` in another terminal. The Go service falls back to `../web/dist` for built static assets when started from `server/`.

## Go Development

```sh
cd server
gofmt -w cmd internal
go build ./...
go run ./cmd/feedback
```

`server/go.mod` uses Go 1.24. The SQLite driver is `modernc.org/sqlite`, so the service builds with `CGO_ENABLED=0`.

Important API behavior:

- `GET /api/health` returns `{"status":"ok"}`.
- `POST /api/auth/login` sets the session cookie.
- `POST /api/auth/register` creates a member account and sets the session cookie. Username and password are required; email and QQ are optional.
- `POST /api/auth/logout` clears the session cookie.
- `GET /api/auth/me` requires a valid session.
- `GET /api/feedback` is public; optional `?category=bug|feature|question` and `?mod=<slug>` filtering are supported. An omitted `mod` uses `rhah`.
- `POST /api/feedback` requires authentication and stores the item on the `mod` query slug.
- `PATCH /api/feedback/{id}` requires the author or administrator session and updates title, body, versions, mod list, and save link.
- `PATCH /api/feedback/{id}/status` lets an author switch only between `open` and `withdrawn`; an administrator can set any valid status.
- `DELETE /api/feedback/{id}` requires the administrator session.
- `GET /api/mods` is public and returns the selectable feedback mods.
- `POST /api/mods`, `PATCH /api/mods/{id}`, and `DELETE /api/mods/{id}` require the administrator session. Deleting a mod also deletes its feedback, and the last mod cannot be deleted.
- `GET /api/settings` is public and returns the displayed mod version, game version, and icon.
- `PUT /api/settings` requires the administrator session.

## Docker Deployment

Create local secrets before the first Compose start:

```sh
mkdir -p secrets
printf '%s' 'admin123456_' > secrets/admin_password
openssl rand -hex 32 > secrets/jwt_secret
docker compose up --build -d
```

The example admin password is `admin123456_`. It must be at least 12 characters. `JWT_SECRET` must be at least 32 characters. The default username is `admin`; override it with `ADMIN_USERNAME`.

Open `http://localhost:8080`. Useful commands:

```sh
docker compose ps
docker compose logs --tail=100 feedback
docker compose down
```

The persistent database is in the `feedback-data` named volume at `/data/feedback.db`. Back it up before upgrades or destructive operations. `docker compose down -v` removes the volume and must be treated as data deletion.

For HTTPS deployments, terminate TLS at a trusted reverse proxy and set `COOKIE_SECURE=true`. Do not expose the service directly to the public Internet without HTTPS and a network-level access policy.

## Docker Desktop + WSL Caveats

Docker Desktop 4.92 was verified with the Linux engine. In this workstation setup, the WSL `archlinux` distro mount service could report:

```text
/run/guest-services/distro-services/archlinux.sock: no such file or directory
```

When that happens, bind-mounted `./secrets` from the WSL project path may fail even though the image builds correctly. Place secret files in a Windows path and set these Compose interpolation variables to the Windows paths:

```powershell
$env:ADMIN_PASSWORD_FILE = 'C:\path\to\admin_password'
$env:JWT_SECRET_FILE = 'C:\path\to\jwt_secret'
```

The Go service reads `ADMIN_PASSWORD_FILE` and `JWT_SECRET_FILE` at startup. Do not put secret contents in `docker-compose.yml` or commit secret files.

The Docker CLI may only be available in PowerShell. The verified executable location was:

```text
C:\Users\miaom\AppData\Local\Programs\DockerDesktop\resources\bin\docker.exe
```

The standalone Compose executable was:

```text
C:\Users\miaom\AppData\Local\Programs\DockerDesktop\resources\bin\docker-compose.exe
```

For a WSL project path, a reliable PowerShell form is:

```powershell
$env:DOCKER_HOST = 'npipe:////./pipe/dockerDesktopLinuxEngine'
& 'C:\Users\miaom\AppData\Local\Programs\DockerDesktop\resources\bin\docker-compose.exe' `
  -f '//wsl.localhost/archlinux/root/repos/golang/modders_feedback/docker-compose.yml' `
  up --build -d
```

If Docker reports `docker-credential-desktop` missing, the Windows Docker config is referencing the Desktop credential helper but the helper is not on the process PATH. Use a temporary Docker config or run from PowerShell with the Docker Desktop installation environment; do not rewrite the user's normal Docker config as a project change.

The image uses public ECR mirrors for Node, Go, and Alpine because Docker Hub and `gcr.io` pulls were unreliable in the verified environment. The final runtime is Alpine 3.22 with UID/GID `65532`, not a root or distroless process.

## Verification Checklist

After code changes:

1. Run `cd web && npm ci && npm run build`.
2. Run `cd server && go build ./...`.
3. Start the service or Compose stack.
4. Check `GET /api/health`.
5. Verify anonymous feedback creation returns `401`.
6. Register a member. Verify the member can create and edit their own feedback, switch it between `open` and `withdrawn`, and receives `403` for another user's feedback, staff statuses, deletion, settings, and mod changes.
7. Log in with the configured administrator account.
8. Update its status and read it back from `GET /api/feedback`.
9. Verify the frontend HTML and its JavaScript asset load.
10. For Docker changes, check `docker compose ps` is `healthy` and inspect logs.

A successful container smoke run produced:

- container state: `Up (healthy)`
- port: `0.0.0.0:8080->8080/tcp`
- health: HTTP 200
- login: HTTP 200
- current user: HTTP 200
- feedback creation: HTTP 201
- feedback listing: HTTP 200

Do not claim Docker validation if only the local Go process or frontend build was exercised.

## Change Guidance

- Keep API and UI field names aligned with the JSON tags in `server/internal/domain/feedback.go` and the `Feedback` type in `web/src/types.ts`.
- If changing an exported or cross-file symbol, inspect all call sites before editing.
- If changing the database schema, provide an additive migration path for existing `/data/feedback.db`; do not silently replace the database.
- If changing auth or cookie behavior, verify login, `me`, logout, expiry, secure-cookie deployment, and rejected unauthenticated writes.
- If changing Docker paths, test both a normal Linux/WSL path and the Docker Desktop Windows/WSL path where available.
- Keep generated `web/dist`, `web/node_modules`, local `data`, `.env`, and `secrets/` out of version control.
