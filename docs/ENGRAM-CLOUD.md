[← Back to README](../README.md)

# Engram Cloud

**Local-first cloud replication for Engram**

<p align="center">
  <img src="../assets/branding/engram-cloud-logo.png" alt="Engram Cloud" width="960" />
</p>

Engram Cloud extends Engram from a single-machine local memory system into a **project-scoped, self-hosted replication layer** with browser visibility for humans and shared continuity across devices.

> Local SQLite remains the source of truth. Cloud is optional replication and shared access — not a replacement for local memory.

---

## What Engram Cloud Is

- A cloud-backed replication layer for Engram
- A browser/dashboard surface for humans
- A self-hosted backend that runs on infrastructure you control
- An explicit, project-scoped sync model

## What Engram Cloud Is Not

- Not cloud-only
- Not a mandatory SaaS
- Not a replacement for local SQLite
- Not an implicit “sync everything” mode

---

## Visual Identity

<p align="center">
  <img src="../assets/branding/engram-cloud-elephant-network.png" alt="Engram Cloud Elephant Network" width="640" />
</p>

Engram Cloud extends the elephant/memory identity of Engram into a distributed cloud/memory-mesh metaphor.

For branding guidance and asset usage, see [Engram Cloud Branding](ENGRAM-CLOUD-BRANDING.md).

---

## Core Principles

### 1. Local-first
SQLite on the developer machine remains authoritative.

### 2. Opt-in cloud
Projects must be explicitly enrolled before cloud sync is allowed.

### 3. Project-scoped replication
Cloud sync is always tied to a single explicit project.

### 4. Graceful degradation
If cloud is unavailable, local Engram must remain usable.

### 5. Self-hosted infrastructure
You choose the server, domain, DNS, reverse proxy, and operational model.

---

## Main Capabilities

### Multi-machine continuity
Carry Engram memory across machines without losing local ownership.

### Browser dashboard
Inspect replicated memory from a browser, including:
- dashboard overview
- browser surfaces
- projects
- contributors
- admin surfaces

### Deterministic failure visibility
Cloud-related failures are surfaced with explicit reason codes instead of silent degradation.

### Guided upgrade path for existing users
Existing local users can migrate via:
- `engram cloud upgrade doctor`
- `engram cloud upgrade repair`
- `engram cloud upgrade bootstrap`
- `engram cloud upgrade status`
- `engram cloud upgrade rollback`

---

## Runtime Surfaces

### Local runtime: `engram serve`
Local JSON API and local `/sync/status` surface.

### Cloud runtime: `engram cloud serve`
Cloud backend with:
- `/health`
- `/sync/pull`
- `/sync/push`
- `/dashboard/*`

### Browser login
In authenticated mode, browser users open `/dashboard/login` and exchange the cloud token for a signed dashboard session cookie.

In insecure local-demo mode (`ENGRAM_CLOUD_INSECURE_NO_AUTH=1`), browser access is simplified for local smoke usage only.

---

## Deployment Shape

Typical production-like shape:

- domain + DNS
- VPS or server you control
- Postgres
- `engram cloud serve`
- reverse proxy / HTTPS

Engram Cloud does **not** require a specific edge stack, but the service expects to sit behind a real HTTP/TLS layer in public deployments.

---

## Client Workflow

### Configure endpoint
```bash
engram cloud config --server https://your-cloud.example.com
```

### Provide token at runtime (authenticated mode)
```bash
export ENGRAM_CLOUD_TOKEN="your-token"
```

### Enroll a project
```bash
engram cloud enroll my-project
```

### Run explicit cloud sync
```bash
engram sync --cloud --project my-project
engram sync --cloud --status --project my-project
```

### Upgrade an existing project before first cloud bootstrap
```bash
engram cloud upgrade doctor --project my-project
engram cloud upgrade repair --project my-project --dry-run
engram cloud upgrade repair --project my-project --apply
engram cloud upgrade bootstrap --project my-project --resume
engram cloud upgrade status --project my-project
```

---

## Deterministic Failure Reasons

Current cloud/runtime surfaces use explicit reason codes such as:

- `blocked_unenrolled`
- `auth_required`
- `cloud_config_error`
- `policy_forbidden`
- `paused`
- `transport_failed`

Upgrade flow also introduces upgrade-specific reason codes and state classes.

---

## Security / Operator Notes

- `ENGRAM_CLOUD_ALLOWED_PROJECTS` is always server-enforced
- authenticated mode requires:
  - `ENGRAM_CLOUD_TOKEN`
  - `ENGRAM_JWT_SECRET` (non-default)
- insecure mode exists only for local smoke/demo use
- browser dashboard auth and sync API auth are intentionally separate concerns

---

## Documentation Map

- [README](../README.md) — product overview + quick start
- [Installation](INSTALLATION.md) — local install methods
- [Agent Setup](AGENT-SETUP.md) — MCP/plugin setup
- [Architecture](ARCHITECTURE.md) — deeper internals
- [Full Docs](../DOCS.md) — technical reference

---

## Note on Visual Branding

If you want to add a dedicated Engram Cloud hero/banner image here, place the image asset inside the repository and link it from this page so the documentation remains self-contained and durable.
