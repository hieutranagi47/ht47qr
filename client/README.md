# Client

Place static client files in `client/api`. The module serves them on the existing
HTTP and HTTPS listeners, directly from the root URL:

- `client/api/index.html` is served at `http://localhost:<SERVER_PORT>/`.
- `client/api/assets/app.js` is served at `http://localhost:<SERVER_PORT>/assets/app.js`.

Only files inside `client/api` are served. Missing files return 404; there is no
SPA fallback. Existing API endpoints retain their paths. The SSE listeners do
not serve client files.

All client GET and HEAD responses include a `Content-Security-Policy` header.
The policy is defined in `client/module.go` and matches the frontend's CSP meta
tag, including inline scripts/styles and `data:`/`blob:` image previews. The
middleware is attached to client routes only.

Files are embedded into the Go binary, so rebuild the service after updating
them. Replace the placeholder `index.html` with your static client build.
