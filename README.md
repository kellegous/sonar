# Sonar

Sonar periodically pings a list of hosts over ICMP, records latency samples in a local [LevelDB](https://github.com/syndtr/goleveldb) store, and serves a web dashboard plus JSON and [Connect](https://connectrpc.com/) RPC APIs.

## Requirements

- [Go](https://go.dev/dl/) 1.26 or newer (see `go.mod`)
- [Bun](https://bun.sh/) — the build uses Bun to install frontend dependencies and run the UI bundle step (`Makefile` targets `node_modules` and `internal/ui/assets`)

ICMP uses raw sockets; on Linux and macOS you typically need elevated privileges (for example `sudo ./bin/sonard …`) or the appropriate capability (for example `CAP_NET_RAW` on Linux).

## Build

From the repository root:

```bash
make
```

This produces `bin/sonard` and embeds the compiled UI under `internal/ui/assets/`. To regenerate protobuf outputs or clean build artifacts, see `Makefile` targets `clean` and `nuke`.

## Run

The server reads a TOML config (default path: `sonar.toml` in the working directory). A sample file is included at the repo root.

```bash
./bin/sonard -conf sonar.toml
```

### Configuration

| Field | Description | Default |
| --- | --- | --- |
| `addr` | HTTP listen address | `:4065` |
| `data-path` | Directory for the LevelDB store (resolved relative to the config file’s directory if not absolute) | `data` |
| `samples-per-period` | ICMP samples taken each period, per host | `10` |
| `sample-period` | Time between sampling rounds ([Go duration](https://pkg.go.dev/time#ParseDuration), e.g. `30s`, `1m`) | `30s` |
| `hosts` | List of `ip` and `name` entries | (none) |

Example:

```toml
addr = ":4065"
data-path = "data"
samples-per-period = 10
sample-period = "1m"

[[hosts]]
ip = "8.8.8.8"
name = "Google"

[[hosts]]
ip = "77.88.8.8"
name = "Russia (Yandex)"

[[hosts]]
ip = "91.239.100.100"
name = "Copenhagen, Denmark"

[[hosts]]
ip = "125.227.80.43"
name = "Beijing, China"
```

## HTTP API

- **Web UI** — `GET /`
- **REST-style JSON**
  - `GET /api/v1/current` — optional query `with-raw=true` to include per-sample RTTs in the response
  - `GET /api/v1/hourly` — optional query `n` (number of hours, default `24`) and `with-raw=true`
- **Connect RPC** — served under `/rpc` (see `sonar.proto` for service and message definitions)

Metrics middleware from [`github.com/kellegous/glue`](https://github.com/kellegous/glue) is attached to HTTP and RPC handlers.

## Development

- **Full stack with live UI** — `make develop` builds `bin/sonard` and runs it with dev mode so assets are served from the Vite dev server (see `cmd/sonard/main.go` and the `develop` target). This target uses `sudo` because raw ICMP and the dev setup expect it.
- **UI only** — from `ui/`: `bun install` (or `npm install` if you mirror scripts) then `bun run dev` / `vite` per `package.json`.

```bash
make test
```

runs Go tests under `internal/...`.

## License

See [LICENSE](LICENSE).
