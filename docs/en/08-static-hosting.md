# Static hosting

Serve HTML, scripts, and images from a directory on the machine that runs the gateway. The same listener can serve APIs and pages: `/api` goes to a backend, `/` or `/assets` reads local files.

Good fit: a Vue / React `dist`, a campaign page, a docs site. Poor fit: putting a whole high-traffic static site on one gateway host (use browser cache or a CDN).

---

## What it does and does not do

**Does**

- Visitors open the gateway URL and get files from a local directory
- Shares the same gateway instance and port as API routes
- Auth / rate limit / CORS on that route still apply
- History-mode frontend routes (`/user/profile` with no file suffix) can fall back to `index.html`

**Does not**

- Directory listing (not a file browser)
- Only GET / HEAD from the browser; no POST upload
- No fetch from object storage (OSS, MinIO); point those routes at a service and proxy
- No sync of directories across hosts; every gateway process must be able to read the path you enter

---

## Two path layers — do not mix them

| You want to change | Where | Does it change the request URL? |
|--------------------|--------|----------------------------------|
| URI seen on the whole chain (logs, later filters, further proxy) | Path rewrite / prefix strip on the route filter | Yes — `Request.URL` |
| Which file on disk | “Map by route path” and “lookup path rewrite” on this page | No — lookup only |
| Path sent to an upstream | Strip prefix / rewrite on the route **proxy** tab | Proxy only; static hosting ignores it |

Lookup is a security boundary: lock the site root, then resolve a relative path inside it, then apply lookup rewrite last. After filters: request path → optional prefix strip → lock the site directory (may include `{v1,v2}`) → optional lookup rewrite (cannot change the root) → file / index; if missing, the same relative path on fallback roots → SPA.

---

## Where to configure

1. Console: **Gateway** → **Routes**
2. Select the gateway instance on the left
3. Add a route (or edit an existing one)
4. Set **backend** to **Static files** (not service proxy)
5. After save, fill in the site directory; or right-click the route → **Static files**

Reload / publish the instance after saving, or the data plane keeps the old config.

A route with neither a service nor a static directory cannot work.

---

## Recommended setups

### 1. Whole frontend at `/`

Build output is `D:/www/shop/dist` (Linux e.g. `/var/www/shop/dist`). Opening `https://gateway/` should be the home page.

1. New route: path `/`, match **prefix**, backend **Static files**
2. Open static hosting and set:

| Field | Suggested value |
|-------|-----------------|
| Local root | `D:/www/shop/dist` |
| Strip route prefix | on |
| SPA fallback | on (Vue/React history mode) |
| Index file | `index.html` |
| Cache seconds | `3600` (js/css for 1 hour; the home page is still not cached) |

`/` or `/index.html` reads `index.html` in that directory. `/assets/app.js` reads `dist/assets/app.js`.

### 2. UI under a subpath, APIs still proxied

`https://gateway/app` is the UI; `https://gateway/api` still goes to a service.

1. API route: path `/api`, prefix, bind a service as usual
2. UI route: path `/app`, prefix, backend **Static files**, then:

| Field | Suggested value |
|-------|-----------------|
| Local root | `D:/www/shop/dist` |
| Strip route prefix | **on** (required) |
| SPA fallback | on |

With strip on, `/app/index.html` reads `index.html` under the root, not `app/index.html`. If strip is off, the gateway looks for `app/index.html` and you get 404s.

Set the frontend base to `/app` as well (e.g. Vite `base: '/app/'`), or asset URLs in the page will be wrong.

### 3. Map old URLs to new files

Example: `/old/a.js` now lives at `new/a.js`. In **lookup path rewrite**, one rule per line, **first match wins**:

```text
prefix /old /new
exact /favicon.ico /images/favicon.ico
regex ^/v\d+/(.*)$ /$1
```

`/old => /new` is the same as the first line.

| Syntax | Meaning | Example |
|--------|---------|---------|
| `prefix /old-prefix /new-prefix` | Replaces only that prefix; longer names are left alone | `/old/a.js` → `new/a.js`; `/older/a.js` unchanged |
| `exact /full-path /target` | Exact path only | `/favicon.ico` → `images/favicon.ico` |
| `regex pattern replacement` | Regex, `$1` for capture groups | `/v2/css/app.css` → `css/app.css` |

Rewrite runs after the directory is locked. In setup 2, `/app/old/a.js` locks `dist`, becomes `/old/a.js`, then reads `new/a.js`. Rules cannot point at another root.

### 4. Several site directories behind one prefix

When directories differ by one name segment (`/var/www/app-v1/dist`, `/var/www/app-v2/dist`), use one **prefix** route plus an allow-list in the root. Do not create a route per directory, and do not put route-regex capture groups into the root path.

1. Route: **prefix**, path e.g. `/apps`
2. Static hosting:

| Field | Suggested value |
|-------|-----------------|
| Site directory | `/var/www/app-{v1,v2,v3}/dist` |
| Map by route path | **on** |
| SPA fallback | depends on the frontend |
| Lookup rewrite | empty |

Request `/apps/v1ui/user/1`: strip prefix → `v1ui/user/1` → first segment locks `/var/www/app-v1/dist` → look up `v1ui/...` inside it. A first segment not on the list is 403. Lookup rewrite cannot switch to `app-v2`. If a directory cannot share the same rule, add another prefix route with a fixed root.

### 5. Login shell + micro-frontends (from nginx)

Keep reverse proxy, method limits, and `%2f` checks on cluster Ingress. 301/302 login entry can be gateway **redirect** routes. The gateway only hosts local directories (e.g. `/data/tomcat`). Static routes use backend **Static files**; methods `GET,HEAD`. Lower priority number matches first; prefix `/` must be the largest number.

**Stay on Ingress**

| nginx | Where |
|-------|--------|
| `/logincenter/datahublogin` → 301 `/#/datahublogin` | Gateway route, backend redirect, 301, target `/#/datahublogin` |
| `/logincenter/datahub_workplace` → 302 | Gateway route, backend redirect, 302, target `/#/datahublogin` |
| `if ($request_method !~* GET\|POST\|HEAD)` | Ingress method limit |
| `if ($request_uri ~ "%2f")` | Ingress reject encoded slash |

Requests map onto directories: turn on “map by route path”; site root uses `{d10,d12}` from the first segment after strip. `/datahub01webVue` and `/datahub01web` must be **regex** so only allowed apps match; a prefix would swallow d01/d05 and unknown paths. No lookup rewrite needed.

**Gateway routes (enter by priority)**

| Pri | Name | Match | Path | Site root | Path map | SPA |
|-----|------|-------|------|-----------|----------|-----|
| 10 | datahub-components | exact | `/lib/sce-vdatahub-components.umd.js` | `/data/tomcat/sce-vdatahub/sce-vdatahub-components` | off | off |
| 20 | vcom-dialogs-lib | regex | `^/lib/sce-vcom-dialogs(?:\.[\w-]+)?\.js$` | `/data/tomcat/sce-vcom/sce-vcom-dialogs` | off | off |
| 30 | vcom-dialogs-umd | prefix | `/sce-vcom-dialogs/umd` | `/data/tomcat/sce-vcom/sce-vcom-dialogs` | on | off |
| 40 | vcom-dialogs | prefix | `/sce-vcom-dialogs` | `/data/tomcat/sce-vcom` | off | off |
| 41 | varea-data | prefix | `/sce-varea-data` | `/data/tomcat/sce-vcom` | off | off |
| 42 | vcom-styles | prefix | `/sce-vcom-styles` | `/data/tomcat/sce-vcom` | off | off |
| 50 | login-assets | prefix | `/assets` | `/data/tomcat/sce-vcom/sce-vcom-login` | off | off |
| 51 | login-lib | prefix | `/lib` | `/data/tomcat/sce-vcom/sce-vcom-login` | off | off |
| 52 | login-images | prefix | `/images` | `/data/tomcat/sce-vcom/sce-vcom-login` | off | off |
| 53 | login-static | prefix | `/static` | `/data/tomcat/sce-vcom/sce-vcom-login` | off | off |
| 54 | login-fonts | prefix | `/fonts` | `/data/tomcat/sce-vcom/sce-vcom-login` | off | off |
| 55 | login-ezuikit | prefix | `/ezuikit` | `/data/tomcat/sce-vcom/sce-vcom-login` | off | off |
| 60 | datahub-vue-d50 | regex | `^/datahub01webVue/(d01\|d05)` | `/data/tomcat/sce-vdatahub/sce-vdatahub-d50-web/umd` | on | on |
| 70 | datahub-vue | regex | `^/datahub01webVue/(d10\|d12\|d13\|d15\|d16\|d18\|d19\|d20\|d21\|d50\|d60\|d70\|d80\|d90)` | `/data/tomcat/sce-vdatahub/sce-vdatahub-{d10,d12,d13,d15,d16,d18,d19,d20,d21,d50,d60,d70,d80,d90}-web/umd` | on | on |
| 80 | datahub-web | regex | `^/datahub01web/(d10\|d12\|d13\|d15\|d16\|d18\|d19\|d20\|d21\|d50\|d60\|d70\|d80\|d90)` | same allow-list | on | on |
| 90 | ai-helper | prefix | `/sceaihelper` | `/data/ai-assistant/ai-frontend/dist` | on | on |
| 100 | a07-web | prefix | `/A07SysBizWebVue` | `/data/tomcat/a07-web/sce-vcom-a07-web/umd` | on | off |
| 110 | versionfiles | prefix | `/versionfiles` | `/data/tomcat` | off | off |
| 120 | login-page | prefix | `/login` | `/data/tomcat/sce-vcom/sce-vcom-login` | on | on |
| 121 | wms-workplace | prefix | `/wms_workplace` | `/data/tomcat/sce-vcom/sce-vcom-login` | on | on |
| 122 | datahub-workplace | prefix | `/datahub_workplace` | `/data/tomcat/sce-vcom/sce-vcom-login` | on | on |
| 130 | login-root | exact | `/` | `/data/tomcat/sce-vcom/sce-vcom-login` | off | off |
| 900 | login-fallback | prefix | `/` | `/data/tomcat/sce-vcom/sce-vcom-login` | off | on |

`/datahub01webVue/d10app/user`: regex hit → strip `/datahub01webVue` → first segment `d10app` expands the root to `sce-vdatahub-d10-web/umd` → read `d10app/user` (or SPA `d10app/index.html`). `/datahub01webVue/other` is not eaten by this route and falls through to the login fallback. d01/d05 are pinned to `d50-web` at priority 60.

Write lookup rewrite only when URL and directory layout disagree: rule 10 `exact /lib/sce-vdatahub-components.umd.js /sce-vdatahub-components.umd.js`, rule 20 `regex ^/lib/(sce-vcom-dialogs(?:\.[\w-]+)?\.js)$ /$1`. Datahub and the login shell need no rewrite. Datahub cache seconds `0`; CORS on the route right-click menu.

**nginx locations that cannot be 1:1**

| nginx | Why | What to do |
|-------|-----|------------|
| `try_files /sec/$app$path /wms/$app$path` | Route captures must not build disk paths | Put the first root in the site directory and the second in fallback roots (same relative path), or split into two prefix routes |
| `root .../sce-vdatahub-$project-web` | Route regex `$1` must not be interpolated into a disk path | Use allow-list `{d10,d12,...}` |
| `add_header X-*` | No arbitrary response headers | Use the security-header allow-list; CORS stays on the route |

---

## Form fields

The form is split into eight groups so lookup, cache, compression, and headers stay separate.

**1. Lookup directories** — site root first; optional fallback roots (same relative path, max 3, no `{v1,v2}`).

**2. Path mapping** — strip prefix, index files, SPA fallback, optional directory slash redirect, exact placeholder match.

**3. Rewrite** — lookup rewrite after the root is locked; cannot change root.

**4. Browser cache** — default max-age for ordinary files; optional per-extension overrides (HTML stays `no-cache`).

**5. Compression** — serve existing `.br`/`.gz` (on by default); optional on-the-fly gzip for text when no precompressed file exists (off by default; skipped for Range and tiny/huge files).

**6. Page security headers** — allow-listed headers only (`X-Frame-Options`, CSP, `Referrer-Policy`, …). Text responses get `charset=utf-8`. CORS stays on the route.

**7. Error pages** — `/404.html`, `/403.html` under the root; status stays 404/403.

**8. Access limits** — extension allow-list, max file size, follow-symlinks.

A directory URL without a trailing `/` is served as an index unless “redirect directory slash” is on. Extension URLs that resolve to a directory stay 404.

Access logs and alerts treat static traffic separately so one page does not dump dozens of js/css rows or flood 404 alerts:

- **Successful hits are not logged**: `file` / `index` / `spa` / `redirect` with status &lt; 400
- **Failures are logged**: 404, 403, 413, 405, 5xx, with `static_result=...; static_path=<disk path>` (the file served, or the last lookup including fallback roots)
- **Alerts**: static 4xx and file-read timeouts do not alert; 5xx still follow instance alert config. Log-write failures still alert.

Change the request URI with route filters; change file mapping with this page. Proxy strip/rewrite is for reverse proxy only. Save validates that regexes compile; if the root already exists it must be a directory.

---

## Security and limits

- Files outside the root are not readable via `../`
- Dot-files are not served by default (`.env`, `.git`); `.well-known` is an exception
- Directory permissions and disk space are an ops concern; the path in the console must be on **the host running the gateway**, not your laptop (unless the gateway runs there)
- Multi-instance or containers: mount the same directory on every instance, or put files on object storage and proxy

---

## Saved but not live?

Check in this order:

1. **Reload / publish** the instance. Console save without reload leaves the old data-plane config.
2. The root exists on the **gateway host**, and the process user can read it.
3. Subpath hosting has “strip route prefix” on.
4. Match type: prefix `/app` covers `/app/xxx`; exact `/app` only equals `/app`.
5. Upgraded databases: if save fails, an admin may need to run the static-hosting table/column patch, then save and reload.

---

## FAQ

| Symptom | What to do |
|---------|------------|
| Blank page or js 404 | Turn on strip prefix for subpath hosting; check frontend `base` vs route path |
| Refresh `/user/1` is 404 | Turn on SPA fallback |
| Requesting a js file returns a full HTML page | Do not rely on SPA for suffixed assets; confirm the file is under the root |
| Home page still old after a release | Expected: home is not long-cached. Hard refresh. Hashed js/css can use a large cache seconds |
| Page should require login | Configure **Auth** on the route; static hosting does not log users in by itself |
| Files are on MinIO / OSS | Do not use static hosting. Add a service node to the store and bind the route |

---

## Static hosting vs binding a service

| Where the files are | How to configure |
|---------------------|------------------|
| Directory on the gateway host | This page: route + static hosting |
| Another HTTP service / nginx / OSS | Route bound to a service (proxy) |
| Just forward a port into the intranet (not site files) | Tunnel static mapping, not this feature |

---

**[Index](./README.md) • [Previous: Error handling](./07-error-handling.md) • [Next: First route](./09-first-route.md)**
