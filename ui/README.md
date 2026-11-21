# Patron UI

React 18 application built with Vite. Use Node 24 (`nvm use` in this directory)
and npm. React Router, the Go API, and runtime authentication are unchanged.

## Development

From `ui/`:

```sh
npm ci
npm start
```

`npm run dev` is an alias for `npm start`. Both use http://localhost:3000 and
fail if that port is occupied. Vite provides Fast Refresh.

The development server serves the repository-root `frontend-config.json` at
`/config.json`, rereading it on every request. Create that file with the installer
or supply the existing deployment settings. The browser needs these keys:

```json
{
  "REACT_APP_NGINX_IP": "localhost",
  "REACT_APP_NGINX_PORT": "8443",
  "REACT_APP_C2SERVER_PORT": "8443"
}
```

Use the actual API hostname and ports for your environment. API requests go
directly to that HTTPS endpoint, so its certificate and CORS settings must permit
the development origin. Missing or invalid configuration returns an explicit HTTP
500 JSON error. After changing configuration, reload the browser; no rebuild is
needed. Configuration is public browser data and must not contain secrets.

## Checks

```sh
npm run lint
npm run test:run
npm run build
```

`npm test` starts Vitest in watch mode. Tests use jsdom and React Testing Library;
API requests and navigation are mocked. ESLint reports existing unused-code,
Hooks, and Fast Refresh findings as warnings; other correctness errors fail.

## Production

`npm run build` emits `build/`, with generated assets under `static/`. The existing
UI Dockerfile installs with `npm ci`, builds with Node 24, and copies this output
to Nginx. Docker Compose mounts `frontend-config.json` read-only at
`/usr/share/nginx/html/config.json`; configuration is never bundled. Nginx keeps
SPA route fallback and returns 404 for missing `/static/` assets.

From the repository root:

```sh
docker build -f Dockerfile.ui -t patron-ui:vite .
docker run --rm -p 8081:8081 \
  --mount type=bind,src="$(pwd)/frontend-config.json",dst=/usr/share/nginx/html/config.json,readonly \
  patron-ui:vite
```

`npm run preview` serves generated files for local inspection only. It does not
serve the repository runtime configuration; use the container workflow above to
check a complete deployment. The development middleware never runs in production.

Check login, authenticated navigation, logout, a refreshed nested route, fonts,
icons, and the browser console against a running API. Check `/config.json` returns
JSON and `/static/missing.js` returns 404 in the container. To roll back, redeploy
the previous UI image; no backend or data migration is required.

Vite's default modern-browser target replaces CRA's Browserslist range. No legacy
browser plugin is included. Builds and linting are separate commands; production
builds do not implicitly run ESLint or tests.
