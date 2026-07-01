# Telescope 🔭

A dynamic reverse proxy written in Go that automatically discovers and routes HTTP traffic to Docker containers — no config files, no restarts required.

## How it works

Telescope watches the Docker event stream. When a container starts with the right labels, it is automatically registered as an upstream route. When it stops, the route is removed. Incoming HTTP requests are forwarded based on the `Host` header.

```
External Client
      │
      │  Host: myapp
      ▼
  Telescope (:8901)
      │
      │  looks up "myapp" in route table
      ▼
  Container (172.17.0.x:8080)
```

## Quick start

```bash
go run .
```

Telescope connects to the local Docker daemon via `DOCKER_HOST` (defaults to the Unix socket at `/var/run/docker.sock`).

## Labelling containers

To register a container with Telescope, add two labels when starting it:

```bash
docker run \
  -l proxy.host=myapp \
  -l proxy.port=8080 \
  my-image
```

Or in a `docker-compose.yml`:

```yaml
services:
  myapp:
    image: my-image
    labels:
      proxy.host: myapp
      proxy.port: 8080
```

| Label | Required | Description |
|---|---|---|
| `proxy.host` | Yes | The `Host` header value Telescope will match incoming requests against |
| `proxy.port` | No | The port your app listens on inside the container. Use if your app exposes multiple ports.  |
| `proxy.path` | No | Optional path prefix to match requests against. When set, only requests starting with this path are routed to the container.|
## Routing requests

Telescope matches the `Host` HTTP header of incoming requests against registered routes.
Be more specific by adding a `proxy.path`, so requests from the same host can match with multiple routes.

In development, add entries to `/etc/hosts` to avoid setting headers manually:

```
127.0.0.1  myapp.local
```

Then set `proxy.host=myapp.local` on your container and make requests normally:

```bash
curl http://myapp.local:8901/
```

Or set the header explicitly:

```bash
curl -H "Host: myapp" http://localhost:8901/
```

## API

### `GET /routes`

Returns the current route table as JSON.

```bash
curl http://localhost:8901/routes
```

```json
[
  { "hostname": "myapp", "address": "172.17.0.3:8080",  path="auth"},
  { "hostname": "api",   "address": "172.17.0.4:3000",  path="api" }
]
```

## Graceful shutdown

Telescope handles `SIGINT` and `SIGTERM`. On shutdown it stops the event watcher, drains in-flight proxy requests, and exits cleanly.

## Project structure

```
.
├── main.go           # Entry point, wires everything together
├── proxy.go          # HTTP reverse proxy handler
├── route_table.go    # RouteTable — thread-safe in-memory route store
├── event_watcher.go  # Event watcher keeps route table up to date
└── docker.go         # Docker client, startup seeding
```