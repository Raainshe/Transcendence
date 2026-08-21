# Transcendence

```bash
make up
```

Open **https://localhost**. The browser will warn about Caddy’s self-signed certificate; proceed anyway (Advanced → Continue / Accept the Risk).

All client-facing backend traffic (REST, public API, swagger, uploads, WebSockets) is HTTPS/WSS on port 443. Port 80 only redirects to HTTPS. Ports 8080 and 5173 are not published.

```bash
make down    # stop containers
make logs    # tail logs
make help    # other targets
```
