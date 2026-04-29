# TinyCMS

A deliberately small Hugo/WordPress-like CMS:

- Go backend using `connect-go` RPC over plain HTTP/JSON.
- SQLite storage with WAL enabled.
- One-user Basic Auth for `/admin/` and API endpoints.
- Public Markdown pages rendered server-side with Goldmark GFM.
- Per-content public routes such as `/`, `/about`, and `/blog/news`.
- Page/post content types plus SEO descriptions for published routes.
- Ant Design React admin UI.
- MDXEditor WYSIWYG-style Markdown editor.
- Optional source Markdown editing mode.
- Links, images, tables, inline code, fenced code blocks, and Mermaid code fences.
- Safe rich Markdown shortcodes for Font Awesome icons and cards.
- File/image uploads inserted into Markdown.
- Global logo, favicon, top menu, public theme, and footer settings.
- IP CIDR allow/deny middleware.
- Optional MaxMind GeoIP country allow/deny middleware.
- Four lightweight admin palettes with `slate` as the default.

## Layout

```text
cmd/tinycms/           server entrypoint
cmsv1connect/          tiny hand-written connect-go bindings using protobuf Struct
internal/auth/         Basic Auth + IP filter middleware
internal/config/       environment config
internal/db/           SQLite store
internal/geo/          optional MaxMind country filter
internal/service/      Connect RPC service implementation
internal/web/          public Markdown renderer
proto/cms.proto        canonical API schema
web/                   React + Ant Design admin
```

## Quick start

```bash
cp .env.example .env
# edit CMS_ADMIN_PASS before exposing the service
make run
```

Open:

- Public site: `http://localhost:8080/`
- Admin: `http://localhost:8080/admin/`

Default login is `admin` / `change-me` unless changed in `.env`.

## Docker

```bash
cp .env.example .env
# edit CMS_ADMIN_PASS
docker compose up --build
```

The Docker build uses the committed `web/package-lock.json` with `npm ci` for repeatable frontend installs.

Use modern Compose (`docker compose`, with a space). The old Python `docker-compose` v1.29.x can fail during container recreation with `KeyError: 'ContainerConfig'` on newer Docker engines. If you hit that, run:

```bash
docker compose up -d --force-recreate --remove-orphans tinycms
```

The Makefile also includes `make docker-up`, `make docker-build`, and `make docker-down` wrappers that use modern Compose.

## Content Model

TinyCMS keeps the editing model intentionally small:

- `Admin slug` is the stable admin identifier used by API calls.
- `Public route / SEO URL` is the published path visitors see, for example `/about/company`.
- `Type` is either `page` or `post`; both render as public Markdown routes for now.
- `SEO description` is emitted as the public page `<meta name="description">`.
- `Published` controls whether the route is visible publicly.

The `home` admin slug is reserved for `/` and cannot be deleted.

## Site Settings

The admin `Site` tab manages simple global pieces shared by every public page:

- Logo and favicon upload fields.
- A nested top menu builder for internal paths such as `/about` and external URLs.
- A Markdown footer shown on every page.
- Public default theme, with a visitor-side light/dark toggle saved in the browser.
- Enable/disable switches for logo, favicon, menu, footer, and the visitor theme toggle.
- Optional Font Awesome loading for the icon shortcode.

The public menu collapses into a small hamburger menu on mobile.

Mermaid diagrams can be written as fenced code blocks:

````markdown
```mermaid
flowchart LR
  A[Idea] --> B[Page]
  B --> C[Publish]
```
````

Small rich-content cards and icons can be written without enabling raw HTML:

```markdown
Build faster {{icon:rocket}} with simple Markdown.

:::card title="Simple operations" icon="gauge-high"
Use **pages**, nested menus, uploads, and Markdown without a heavy page builder.
:::
```

Icon names map to Font Awesome solid classes, so `{{icon:rocket}}` becomes `fa-solid fa-rocket`. Advanced classes such as `{{icon:fa-brands fa-github}}` are also accepted, but only letters, numbers, spaces, and dashes are allowed.

## Config

| Variable | Default | Notes |
|---|---:|---|
| `CMS_ADDR` | `:8080` | Listen address. |
| `CMS_SITE_NAME` | `TinyCMS` | Public site name. |
| `CMS_ADMIN_USER` | `admin` | Basic Auth username. |
| `CMS_ADMIN_PASS` | `change-me` | Basic Auth password. Change this. |
| `CMS_DATA_DIR` | `./data` | Data root. |
| `CMS_DB` | `./data/cms.db` | SQLite DB path. |
| `CMS_UPLOAD_DIR` | `./data/uploads` | Upload directory. |
| `CMS_MAX_UPLOAD_BYTES` | `26214400` | Max upload size. |
| `CMS_ALLOW_CIDRS` | empty | Comma-separated CIDRs. Empty means allow all. |
| `CMS_DENY_CIDRS` | empty | Comma-separated denied CIDRs. |
| `CMS_TRUST_PROXY_HEADERS` | `false` | Enables `CF-Connecting-IP`, `X-Real-IP`, and `X-Forwarded-For`. Only enable behind trusted reverse proxy. |
| `CMS_MAXMIND_DB` | empty | Path to GeoLite2/GeoIP2 Country `.mmdb`. Empty disables geo filtering. |
| `CMS_ALLOW_COUNTRIES` | empty | ISO country allow list, e.g. `US,CA`. Empty means allow all except denied. |
| `CMS_DENY_COUNTRIES` | empty | ISO country deny list. |

## Security notes

- Put it behind TLS. Basic Auth is only safe over HTTPS.
- Set a strong `CMS_ADMIN_PASS`.
- Keep `CMS_TRUST_PROXY_HEADERS=false` unless a trusted proxy strips and rewrites those headers.
- Public uploads are limited to common image/text/document extensions.
- Raw HTML in Markdown is escaped by default; use Markdown syntax for page content.

## API shape

This project uses Connect RPC endpoints but keeps the code minimal by using `google.protobuf.Struct` request/response payloads instead of generated app-specific message structs.

Example:

```bash
curl -u admin:change-me \
  -H 'Content-Type: application/json' \
  -d '{"slug":"home"}' \
  http://localhost:8080/cms.v1.CMSService/GetPage
```

## Why this design

This intentionally avoids a plugin system, themes marketplace, server-side JS, full-text search, users/roles, and complex media workflows. The core path is a few SQL queries, Markdown rendering, static upload serving, and a compact React admin.
