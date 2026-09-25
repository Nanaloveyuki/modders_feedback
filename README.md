# Ratkin: Hunger and Havoc Feedback

A self-hosted feedback board for the RimWorld 1.6 mod **Ratkin: Hunger and Havoc** (`nanaloveyuki.ratkin.hungerandhavoc`). Go serves the API and built React app; SQLite stores accounts and feedback in a persistent Docker volume.

## Start with Docker Compose

Requirements: Docker Engine with the Compose plugin.

For first run, create two local secret files (the `secrets/` directory is ignored by Git):

```sh
mkdir -p secrets
printf '%s' 'admin123456_' > secrets/admin_password
openssl rand -hex 32 > secrets/jwt_secret
docker compose up --build -d
```

Copy or replace the example value in `secrets/admin_password` before exposing the site; it must contain at least 12 characters. `jwt_secret` must contain at least 32 characters. Never commit the secret files. Set `FEEDBACK_PORT` to change the published port; `ADMIN_USERNAME` defaults to `admin` and the example password is `admin123456_`.

Open `http://localhost:8080`. The initial administrator account is seeded on the first start. For internet-facing deployments, terminate HTTPS at a trusted reverse proxy and set `COOKIE_SECURE=true`.

The image uses Alpine runtime and public ECR mirrors for the Node, Go, and Alpine base images. This avoids Docker Hub and gcr.io registry issues in restricted networks. On Docker Desktop with WSL, the Compose project path must be accessible to Docker Desktop; if WSL secret mounts report a missing `archlinux.sock`, place the two secret files under a Windows path and set `ADMIN_PASSWORD_FILE` and `JWT_SECRET_FILE` to those Windows paths before starting. Docker Desktop 4.92 was verified with copied Windows secret files; the container reported `healthy` and served port `8080`.

Change the published port with `FEEDBACK_PORT`. To stop without deleting feedback:

```sh
docker compose down
```

`docker compose down -v` deletes the feedback database; use it only when intentionally discarding all stored data.

## Feedback workflow

The board has bug reports, feature requests, and general questions. Signed-in users can create feedback; visitors can read and filter posts. Bug reports include optional RimWorld version, mod version, mod list/load order, and save-share URL fields. The account supplied in the environment is the only account; registration is disabled. A signed-in administrator can update feedback status, edit title, body, versions, mod list, and save link, delete a record, and change the displayed mod version, game version, and site icon.

SQLite is created at `/data/feedback.db`; the Compose configuration persists `/data` in the `feedback-data` volume. The service also writes an online backup to `/data/backups/feedback-<UTC timestamp>.db` at startup and every 72 hours. It keeps the 10 newest completed copies and deletes the oldest when a new copy is published. If the disk is full, it deletes the oldest completed copy and retries once; a failed attempt is removed and does not replace a good copy. A backup panic, including an out-of-memory failure inside the backup, is logged and does not stop the HTTP server. `GET /api/health` is available for monitoring.

## Local development

The web app runs with Vite and proxies `/api` to the Go service during development. Build the frontend with `cd web && npm ci && npm run build`. Start the API from the `server/` directory with `DATA_DIR=../data ADMIN_USERNAME=admin ADMIN_PASSWORD='admin123456_' JWT_SECRET='local-random-secret-at-least-32-characters' go run ./cmd/feedback`; it serves the built `web/dist` files when started there. For Vite hot reload, run `npm run dev` from `web/` in a second terminal.
