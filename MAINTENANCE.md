# OpenGist Fork Maintenance Plan

## Overview
Custom fork of [thomiceli/opengist](https://github.com/thomiceli/opengist) with two patches:
1. **Everyone-can-edit** — any logged-in user can edit any gist (with proper git attribution)
2. **Cloudflare Access auto-login** — automatic SSO login via `Cf-Access-Authenticated-User-Email` header

## Repositories
- **Upstream:** `git@github.com:thomiceli/opengist.git` (remote: `origin`)
- **Fork:** `git@github.com:potatosalad/opengist.git` (remote: `potatosalad`)
- **Local clone:** `~/Work/maclan/opengist`
- **SSH key:** `$HOME/.ssh/github.com_id_ed25519`
- **Docker registry:** `192.168.86.209:5050`

## Patches (on branch `everyone-can-edit-YYYY-MM-DD`)

### Patch 1: everyone-can-edit
- **What it does:** Any authenticated user can edit any gist. Edits are attributed to the editing user in the git commit (author field), preserving the original gist owner. Delete and visibility changes remain owner-only.
- **Files modified:**
  - `internal/web/server/router.go` — removed `writePermission` middleware from edit routes
  - `internal/db/gist.go` — commit functions accept editor's name/email
  - `internal/web/handlers/gist/create.go` + `edit.go` — pass editing user info to commits
  - `templates/base/gist_header.html` — Edit button visible to all logged-in users

### Patch 2: Cloudflare Access auto-login
- **What it does:** When the `Cf-Access-Authenticated-User-Email` HTTP header is present, looks up the user by email and auto-creates a session. Only works for existing accounts; skips if already logged in.
- **Files modified:**
  - `internal/db/user.go` — added `GetUserByEmail()` function
  - `internal/web/server/middlewares.go` — `cfAccessAutoLogin` middleware after `sessionInit`

### Initial branch: `everyone-can-edit-2026-03-01`

## Monthly Maintenance Procedure

### 1. Sync upstream
```bash
cd ~/Work/maclan/opengist
git fetch origin
git checkout master
git pull origin master
git push potatosalad master
```

### 2. Rebase patch branch
```bash
git checkout everyone-can-edit-2026-03-01
git rebase master
# Resolve conflicts if any
```

### 3. Build & test
```bash
# Build for amd64 (we're on arm64)
docker buildx build --platform linux/amd64 -t 192.168.86.209:5050/opengist:latest -f Dockerfile .

# Quick smoke test
docker run --rm -p 6157:6157 192.168.86.209:5050/opengist:latest
# Verify:
# - Create 2+ users, test cross-user gist editing, check git log attribution
# - Test Cloudflare Access header: curl -H "Cf-Access-Authenticated-User-Email: user@example.com" http://localhost:6157/
#   (should auto-login if user with that email exists)
# - Confirm delete button is still owner-only
```

### 4. Push
```bash
git push potatosalad everyone-can-edit-2026-03-01 --force-with-lease
docker push 192.168.86.209:5050/opengist:latest
```

### 5. Report
Send a summary to Andrew via Telegram:
- Upstream changes since last sync
- Whether rebase was clean or had conflicts
- Build status
- Any breaking changes affecting either patch

## If Something Breaks
- Check upstream CHANGELOG for breaking changes
- Compare diff between old and new upstream on files we patched
- Key conflict areas: `router.go` (middleware chain), `middlewares.go` (session handling), `gist.go` (commit API), `user.go` (user model), `gist_header.html` (template)
- If patch no longer applies cleanly, create a new dated branch and re-apply manually
