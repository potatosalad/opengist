# OpenGist Fork Maintenance Plan

## Overview
Custom fork of [thomiceli/opengist](https://github.com/thomiceli/opengist) with five patches:
1. **Everyone-can-manage** — any logged-in user can edit, delete, and change visibility of any gist
2. **Cloudflare Access auto-login** — automatic SSO login via `Cf-Access-Authenticated-User-Email` header
3. **Markdown relative image resolution** — relative image refs in markdown files resolve to sibling files in the same gist
4. **Collapsible file sections** — binary files (images, audio, video, PDF) are collapsed by default; all files can be toggled
5. **Markdown files first** — markdown files are listed before all other files on the gist display page

## Repositories
- **Upstream:** `git@github.com:thomiceli/opengist.git` (remote: `origin`)
- **Fork:** `git@github.com:potatosalad/opengist.git` (remote: `potatosalad`)
- **Local clone:** `~/Work/maclan/opengist`
- **SSH key:** `$HOME/.ssh/github.com_id_ed25519`
- **Docker registry:** `192.168.86.209:5050`

## Patches (on branch `everyone-can-edit-YYYY-MM-DD`)

### Patch 1: everyone-can-manage
- **What it does:** Any authenticated user can edit, delete, and change visibility (public/unlisted/secret) of any gist. Edits are attributed to the editing user in the git commit (author field), preserving the original gist owner.
- **Files modified:**
  - `internal/web/server/router.go` — removed `writePermission` middleware from edit, visibility, and delete routes
  - `internal/db/gist.go` — commit functions accept editor's name/email
  - `internal/web/handlers/gist/create.go` + `edit.go` — pass editing user info to commits
  - `templates/base/gist_header.html` — Edit and Delete buttons visible to all logged-in users (owner-only gate removed from delete)

### Patch 2: Cloudflare Access auto-login
- **What it does:** When the `Cf-Access-Authenticated-User-Email` HTTP header is present, looks up the user by email and auto-creates a session. Only works for existing accounts; skips if already logged in.
- **Files modified:**
  - `internal/db/user.go` — added `GetUserByEmail()` function
  - `internal/web/server/middlewares.go` — `cfAccessAutoLogin` middleware after `sessionInit`

### Patch 3: Markdown relative image resolution
- **What it does:** When a gist contains a markdown file (e.g. `doc.md`) alongside image files (e.g. `photo.jpg`), relative image references like `![](photo.jpg)` or `![](./photo.jpg)` are rewritten to point to the gist's raw file endpoint, so images render inline correctly.
- **Files modified:**
  - `internal/render/markdown_relative_links.go` — new goldmark AST transformer `relativeImageRewriter` that rewrites relative image destinations to `{baseURL}/{user}/{gist}/raw/{revision}/{filename}`
  - `internal/render/markdown.go` — `renderMarkdownFile` accepts `rawBaseURL`; `newMarkdown()` conditionally includes the transformer
  - `internal/render/render.go` — `RenderFiles` and `processFile` accept `rawBaseURL` and pass it to markdown rendering
  - `internal/web/handlers/gist/gist.go` — `rawBaseURL()` helper builds the base URL from context; all three `RenderFiles` call sites (`GistIndex`, `GistJson`, `GistJs`) pass the raw base URL

### Patch 4: Collapsible file sections
- **What it does:** File content sections on the gist view page can be collapsed/expanded by clicking the file header. Binary files (images, audio, video, PDF) start collapsed by default; text files start expanded. A global "Collapse all / Expand all" toggle is available at the top of the file list.
- **Files modified:**
  - `templates/pages/gist.html` — added `data-binary` attribute to file cards, chevron icon in file headers, `gist-file-toggle`/`gist-file-body`/`gist-file-meta` CSS classes, and "Expand/Collapse all" button
  - `public/ts/gist.ts` — collapse/expand logic: `setFileCollapsed()`, `isCollapsed()`, `updateToggleAllButton()`, per-file click handlers, global toggle button handler

### Patch 5: Markdown files first
- **What it does:** On the gist display page, markdown files (.md) are sorted alphabetically before all other file types. Non-markdown files follow in their own alphabetical order. This ensures documentation is always visible at the top.
- **Files modified:**
  - `internal/db/gist.go` — added `slices.SortStableFunc` in `Files()` method to sort markdown-extension files before others

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
# - Create 2+ users, test cross-user gist editing, deleting, and visibility changes
# - Check git log attribution (edits attributed to editing user, not owner)
# - Test Cloudflare Access header: curl -H "Cf-Access-Authenticated-User-Email: user@example.com" http://localhost:6157/
#   (should auto-login if user with that email exists)
# - Create a gist with a .md file + image file, verify ![](image.jpg) renders inline
# - Verify binary files are collapsed by default, text files expanded
# - Test per-file toggle (click header) and global "Collapse/Expand all" button
# - Create a gist with a .md file + non-md files, verify .md appears first
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
- Any breaking changes affecting any of the four patches

## If Something Breaks
- Check upstream CHANGELOG for breaking changes
- Compare diff between old and new upstream on files we patched
- Key conflict areas: `gist.go` (file sorting/commits), `router.go` (middleware chain), `middlewares.go` (session handling), `gist.go` (commit API), `user.go` (user model), `gist_header.html` (template), `gist.html` (view template), `gist.ts` (client JS), `render.go` / `markdown.go` (rendering pipeline)
- If patch no longer applies cleanly, create a new dated branch and re-apply manually
