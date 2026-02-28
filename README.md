# dogowaiter

Docker cross-stack dependency waiter in Go. Project layout follows [golang-standards/project-layout](https://github.com/golang-standards/project-layout): `main.go` (entrypoint), `cmd/dogowaiter` (command), `internal/dogowaiter` (app logic), `build/package` (Docker).

## Docker container usage

Run the image (mount the Docker socket; the container needs it to talk to the daemon):

```bash
docker run --rm -it \
  -v /var/run/docker.sock:/var/run/docker.sock \
  dogowaiter [options]
```

See [Usage](#usage) for CLI flags and env vars.

**Example `compose.yaml`** with all env vars:

```yaml
services:
  dogowaiter:
    image: ghcr.io/nekrich/dogowaiter:latest
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./config.yaml:/config/dogowaiter.yaml:ro  # optional, for --config
    environment:
      DOGOWAITER_CONFIG_FILE: /config/dogowaiter.yaml      # optional; or use DOGOWAITER_DEPENDENCIES
      DOGOWAITER_DEPENDENCIES: "mystack:web,api"           # optional; used when no config file
      DOGOWAITER_HEALTH_FILE: /tmp/healthy                 # optional;
      DOGOWAITER_DOCKER_HOST: unix:///var/run/docker.sock  # optional
      DOGOWAITER_LOG_LEVEL: info                           # optional
    # Or use DOCKER_HOST / LOG_LEVEL instead of DOGOWAITER_* where supported
```

## Usage

CLI flags and env vars (flags override env). At least one of `--config` / DOGOWAITER_CONFIG_FILE or `--dependencies` / DOGOWAITER_DEPENDENCIES is required.

| Flag | Env | Description |
|------|-----|-------------|
| `--config` | `DOGOWAITER_CONFIG_FILE` | Path to YAML file with a `depends_on` list. If empty, `DOGOWAITER_DEPENDENCIES` is used instead. When set, the file is watched (resolved path, 30s debounce); changes are picked up automatically without restart. If the file is removed, the process resolves the path once more and exits with error status if it no longer exists. Default: `config/dogowaiter.yaml`. |
| `--dependencies` | `DOGOWAITER_DEPENDENCIES` | Comma-separated list of dependencies. Each entry is `stack:service` or `service` (stack omitted). Examples: `stack1:web,stack2:api`, `web,api`, `service` (empty stack). Used when config file is not set. |
| `--docker-host` | `DOGOWAITER_DOCKER_HOST`, `DOCKER_HOST` | Docker API endpoint. Default: `unix:///var/run/docker.sock`. For docker-socket-proxy or remote Docker, set to e.g. `tcp://docker-socket-proxy:2375`. |
| `--log-level` | `DOGOWAITER_LOG_LEVEL`, `LOG_LEVEL` | Log level: `debug`, `info`, or `error`. Default: `info`. |
| | `DOGOWAITER_HEALTH_FILE` | Path for the JSON health file. Default: `/tmp/healthy`. |

## Config format (YAML)

When using **DOGOWAITER_CONFIG_FILE**, the file can contain:

```yaml
docker_host: tcp://sock:2375   # optional
log_level: debug               # optional

depends_on:
  - name: myservice
    stack: mystack             # optional; empty = any stack. Match by com.docker.compose.project label.
  - otherservice               # scalar = service name only (any stack)
```

Containers are matched by label `com.docker.compose.project` (when stack is set) and `com.docker.compose.service` or by container name.

## Health file

- **healthy**: boolean; true when all monitored containers are ready.
- **containers**: array of all monitored containers; each has `container`, `reason`, `is_ready`. Ready is derived automatically: running and (no health config or health status == "healthy").

Example: `{"healthy": false, "containers": [{"container": "web", "is_ready": false, "reason": "unhealthy"}]}`.

## Dockerfile HEALTHCHECK

The image HEALTHCHECK runs `grep -q '"healthy": true' /tmp/healthy`.

## Build & run locally

From repo root. Plain `go build` does nothing here (no Go files at root); use the package path or [mise](https://mise.jdx.dev/) file tasks:

```bash
# With mise (recommended)
mise go:build    # → binary ./dogowaiter
mise go:test
mise go:run      # build and run in one step

# Without mise
go build -o dogowaiter .
./dogowaiter
```

Docker (from repo root):

```bash
mise docker:build
mise docker:run
```

It reads config (YAML file or env), checks that listed services (containers) from other stacks are running/healthy via the Docker API, subscribes to Docker Events, and writes a JSON health file, so you can easily depend on healthy dogowaiter to run your cross-stack services.

## CI

- **test** (`.github/workflows/test.yaml`): on push/PR to main — `mise run go:test`.
- **publish** (`.github/workflows/publish.yaml`): on release, tag push `v*`, or manual — `mise run docker:build`, then publish to GHCR and Docker Hub via `mise run image:publish_github` and `mise run image:publish_dockerhub`. Required secrets: **DOCKERHUB_USERNAME**, **DOCKERHUB_TOKEN**. GHCR uses `GITHUB_TOKEN`.
