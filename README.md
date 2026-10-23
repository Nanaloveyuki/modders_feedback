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

The image builds with Node 24.22.1 and Go 1.26.10 on Alpine 3.24, then runs on Alpine 3.24.3 as UID/GID `65532`. It uses public ECR mirrors to avoid Docker Hub and gcr.io registry issues in restricted networks. On Docker Desktop with WSL, the Compose project path must be accessible to Docker Desktop; if WSL secret mounts report a missing `archlinux.sock`, place the two secret files under a Windows path and set `ADMIN_PASSWORD_FILE` and `JWT_SECRET_FILE` to those Windows paths before starting. The earlier Docker Desktop 4.92 deployment was verified with copied Windows secret files; that result does not validate the updated images.

Change the published port with `FEEDBACK_PORT`. To stop without deleting feedback:

```sh
docker compose down
```

`docker compose down -v` deletes the feedback database; use it only when intentionally discarding all stored data.

## Feedback workflow

The board has bug reports, feature requests, and general questions. Visitors can read and filter posts or register with a username, password, and optional email or QQ number. Members can create feedback, edit their own content, and switch between author-switchable statuses while their current status permits it. Administrators can update any status, edit or delete feedback, and manage mods and site settings. Bug reports include optional RimWorld version, mod version, mod list/load order, and save-share URL fields. Create and edit forms support image paste and PNG, JPEG, GIF, WebP, PDF, and plain-text attachments. Files use fast zstd compression by default; statuses marked archived trigger maximum compression in the background.

SQLite is created at `/data/feedback.db`; the Compose configuration persists `/data` in the `feedback-data` volume. The service also writes an online backup to `/data/backups/feedback-<UTC timestamp>.db` at startup and every 72 hours. It keeps the 10 newest completed copies and deletes the oldest when a new copy is published. If the disk is full, it deletes the oldest completed copy and retries once; a failed attempt is removed and does not replace a good copy. A backup panic, including an out-of-memory failure inside the backup, is logged and does not stop the HTTP server. `GET /api/health` is available for monitoring.

Back up the database and attachment directory before upgrading; startup migrations run before the automatic startup backup. Status migrations preserve attachments and timelines when rebuilding feedback tables and no longer repeat on an already-migrated schema. A database backup alone does not include uploaded files.

`GET /rss` is a public RSS 2.0 feed of the 50 newest feedback items across mods. Each item title is `mod-category-title`, and the description is the written feedback text. Image, video, and file markup is removed. `Accept-Language: en` switches the feed language and category labels to English.

## MCP for local OMP

The existing Go process serves Streamable HTTP at `/mcp` using the official Go MCP SDK. Anonymous tools: `list_mods`, `list_feedback`, `get_feedback`, `list_feedback_timeline`, and `get_catalog`. Authenticated tools: `whoami`, `update_feedback`, `set_feedback_status`, `create_mod`, and `update_mod`. No delete, upload, account-management, or server-settings tools are exposed. See [the accepted decision](docs/adr/0001-local-omp-mcp.md).

Lists accept `limit` (default 20, maximum 100) and `offset` (0–10000); `nextOffset` means another page may exist. `list_mods` accepts a name/slug substring in `query`. `list_feedback` requires a `mod` slug and accepts `category` and configured `status` filters. Categories remain `bug`, `feature`, and `question`; use `get_catalog` for current statuses and permitted icons. Detail, timeline, and feedback writes use `publicId`; mod edits use the stable numeric `id`. Returned web and attachment paths are relative to the site origin.

### Enable writes

Without MCP credentials configured, only public reads work. To enable writes, associate one random token with an existing account. Generate and retain the token on the client; copy only its digest to the server:

```sh
MCP_TOKEN=$(openssl rand -hex 32)
printf '%s' "$MCP_TOKEN" | sha256sum
export MODDERS_FEEDBACK_MCP_AUTH="Bearer $MCP_TOKEN"
unset MCP_TOKEN
```

Set server environment `MCP_USERNAME` to the existing account name and `MCP_TOKEN_SHA256` to the 64-character digest, then restart. Compose passes both variables through; direct service configuration also supports `MCP_USERNAME_FILE` and `MCP_TOKEN_SHA256_FILE`. Configure both or neither. Replacing the digest and restarting rotates the token; clearing both disables writes. Removing or renaming the associated account invalidates its token. Account roles are read on every authenticated request.

The token does not grant extra permissions: members can edit only their own feedback and can switch status only when both current and target labels allow authors. Administrators can edit any feedback and create/edit mods. Feedback and mod edits accept partial fields; omission preserves the current value, and an empty optional string clears it. Empty titles/bodies, empty patches, invalid values, and unknown fields are rejected. For example, `update_feedback` accepts `{"publicId":"<feedback UUID>","title":"Corrected report title"}`; `update_mod` accepts `{"id":2,"githubUrl":""}`.

### Connect OMP

Merge this entry into your local OMP `mcp.json` without committing credentials. Start OMP from a shell containing `MODDERS_FEEDBACK_MCP_AUTH`; its value must include the `Bearer ` prefix. The environment-variable-name header resolution was verified with the installed OMP client. Replace the URL with the actual HTTPS site URL for remote access:

```json
{
  "mcpServers": {
    "modders-feedback": {
      "type": "http",
      "url": "http://127.0.0.1:8080/mcp",
      "headers": { "Authorization": "MODDERS_FEEDBACK_MCP_AUTH" }
    }
  }
}
```

For anonymous reads, omit `headers` entirely. An invalid explicit token returns 401 even for public tools; website cookies never authenticate MCP. This is a dedicated bearer-token integration, not OAuth auto-login or a public-key system. Use HTTPS remotely or a trusted local tunnel. A local OMP client does not make `/mcp` network-private: the endpoint shares the website listener, and public reads remain reachable wherever the site is reachable.

Verification (2026-10-23): real Go SDK and installed OMP HTTP clients initialized and called tools; isolated regression tests covered member/admin permissions, configurable statuses, partial edits preserving attachments, rejected unknown fields, and live role lookup. Live restart checks covered file-based credentials, token rotation (old token rejected, new token accepted), and persisted feedback. Oversized bodies, batches, and untrusted origins were rejected. Website HTTP checks covered login, registration, feedback creation/editing, status changes, denied privileged actions, and frontend HTML/JavaScript loading. Frontend production build, CGO-free backend build, and `go test -race ./...` passed. Windows Compose configuration validation passed; no production deployment or Docker Compose runtime verification was performed for this change.

## Local development

Requirements: Go 1.26 or newer and Node.js 22.12+ or 24+; Node 24.22.1 is the verified container build version. The frontend uses React 19.3, Vite 8, TypeScript 7, and Lucide 1.61. Build tooling belongs in `devDependencies`; deployment builds must install development dependencies before bundling.

The web app runs with Vite and proxies `/api` to the Go service during development. Build the frontend with `cd web && npm ci && npm run build`. Start the API from the `server/` directory with `DATA_DIR=../data ADMIN_USERNAME=admin ADMIN_PASSWORD='admin123456_' JWT_SECRET='local-random-secret-at-least-32-characters' go run ./cmd/feedback`; it serves the built `web/dist` files when started there. For Vite hot reload, run `npm run dev` from `web/` in a second terminal.

### Dependency update verification (2026-10-23)

- `npm ci` and the frontend production build passed, including inside the Node 24.22.1 Alpine image filesystem. `npm audit` reported zero vulnerabilities.
- Go 1.26.10 built the CGO-free service and passed `go test ./...`. `govulncheck` found no reachable vulnerabilities; module-level scanning still reports GO-2026-5932 in the unused `golang.org/x/crypto/openpgp` package, with no fixed version available. The service uses `bcrypt`, not OpenPGP.
- Removed the untrusted-header `RealIP` middleware. Restored the missing feedback-edit route and added author/admin permission regression coverage.
- Real HTTP and browser checks covered authentication, permissions, feedback creation/editing, comments, attachment access and archive compression, WebP avatars, Markdown sanitization, RSS, desktop/mobile layouts, restart preservation, and backup restore using temporary data.
- The service also ran with UID/GID `65532`, a read-only Alpine 3.24.3 root, and dropped capabilities. This was an isolated filesystem smoke run, not a Docker Compose deployment verification.


### OMP language servers

`.omp/lsp.json` enables Go, TypeScript/React, HTML, CSS, and JSON language servers when OMP starts at the repository root, despite the nested `server/` and `web/` projects. Install Go tooling with `go install golang.org/x/tools/gopls@latest` and HTML/CSS/JSON tooling with `npm install -g vscode-langservers-extracted`; ensure `gopls` is on OMP's PATH. Run `npm ci` in `web/` to provide the project's TypeScript 7 native server (`tsc --lsp --stdio`); no separate `typescript-language-server` is needed.

The HTML/CSS/JSON command paths currently match this workstation's `/usr/sbin/` global npm binaries. Adjust those paths for another installation. After installation or configuration changes, use OMP's LSP tool with `{"action":"reload","file":"*"}`, then request hover or definitions on a source file to verify the server is working.
