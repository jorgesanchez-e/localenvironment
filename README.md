# localenvironment

A monorepo of small Go programs used to run a local or self-hosted environment: one application manages **Dynamic DNS (DDNS)** records, and another handles **TLS certificates** issued through the **ACME** protocol (for example, Let's Encrypt).

Together, these tools help keep hostnames pointing at the right address and keep HTTPS certificates renewed without manual intervention.

## Applications

| Application | Path | Purpose |
|-------------|------|---------|
| **ddns** | [`apps/ddns`](apps/ddns) | Update DNS records when your public IP or target address changes. |
| **acme** | [`apps/acme`](apps/acme) | Obtain and renew certificates using the ACME protocol. |

Each app lives under `apps/<name>/` with its own `go.mod`, Makefile, and `scripts/` helpers.

## Prerequisites

- [Go](https://go.dev/) (see `go` in `apps/ddns/go.mod` or `apps/acme/go.mod`)
- [Git](https://git-scm.com/) (recommended; used to resolve the repository root)
- **make** — GNU Make or BSD Make (macOS, Linux, FreeBSD)
- **curl** — only required for `make lint` (installs [golangci-lint](https://golangci-lint.run/))

Build and test scripts use POSIX `/bin/sh` and are intended to work on **macOS**, **Linux**, and **FreeBSD**.

## Repository layout

```
.
├── Makefile          # Build all apps from the repo root
├── build/            # Compiled binaries (created by `make build`)
└── apps/
    ├── ddns/
    └── acme/
```

## Building binaries

### Build everything (recommended)

From the repository root:

```bash
make
# or
make build
```

This discovers every app under `apps/*/Makefile` (currently `ddns` and `acme`) and builds each one.

Binaries are written to the shared `build/` directory at the repo root, with names like:

```text
build/ddns-<GOOS>-<GOARCH>-<APPVERSION>.bin
build/acme-<GOOS>-<GOARCH>-<APPVERSION>.bin
```

Example on macOS arm64 with the default version `0.1`:

```text
build/ddns-darwin-arm64-0.1.bin
build/acme-darwin-arm64-0.1.bin
```

### Build a single application

```bash
make -C apps/ddns build
make -C apps/acme build
```

Or from inside an app directory:

```bash
cd apps/ddns
make build
```

### Cross-compilation and version

Pass variables from the root or from an app Makefile; they are forwarded to the build script:

| Variable | Default | Description |
|----------|---------|-------------|
| `GOOS` | Host OS (`go env GOOS`) | Target operating system (`linux`, `freebsd`, `darwin`, …) |
| `GOARCH` | Host arch (`go env GOARCH`) | Target architecture (`amd64`, `arm64`, …) |
| `APPVERSION` | `0.1` | Version segment in the output filename |
| `ROOT_DIR` | Git root | Repository root (artifacts go under `$(ROOT_DIR)/build`) |

Examples:

```bash
# Linux amd64 binaries from any host
make build GOOS=linux GOARCH=amd64 APPVERSION=1.0.0

# Build only ddns for FreeBSD arm64
make -C apps/ddns build GOOS=freebsd GOARCH=arm64
```

Builds use `CGO_ENABLED=0` for static, portable binaries.

### Run a built binary

After building, run an app with:

```bash
make -C apps/ddns run
make -C apps/acme run
```

This executes the matching binary from `build/` (same naming rules as above).

## Other Make targets

### At the repository root

| Target | Description |
|--------|-------------|
| `make clean` | Remove build artifacts for all apps from `build/` |
| `make test` | Run unit tests for all apps |
| `make lint` | Run golangci-lint for all apps |
| `make test-report` | Open HTML coverage report (requires prior `make test`) |

### Per application (`apps/ddns` or `apps/acme`)

| Target | Description |
|--------|-------------|
| `make build` | Compile the app binary into `build/` |
| `make run` | Build (if needed) and run the binary |
| `make test` | Run tests with race detector and coverage |
| `make lint` | Lint with golangci-lint (version from `GOLANGCILINT_VERSION` in the app Makefile) |
| `make clean` | Remove this app’s binaries from `build/` |
| `make test-report` | Show coverage HTML for this app |

Test coverage files are stored under `apps/<name>/reports/`.

## Go workspace (optional, local)

`go.work` is listed in [`.gitignore`](.gitignore) and is not required for `make build` or `make test`. To use a workspace locally, create it once at the repo root:

```bash
go work init ./apps/ddns ./apps/acme
go work sync   # optional: refresh workspace dependencies
```

Keep the `go` version in each app’s `go.mod` in sync; CI reads the toolchain from [`apps/ddns/go.mod`](apps/ddns/go.mod).

## Continuous integration

GitHub Actions runs on every **pull request** and on pushes to `main` / `master`. The workflow is defined in [`.github/workflows/ci.yml`](.github/workflows/ci.yml) and runs two jobs on Ubuntu in parallel:

| Job | Command | Description |
|-----|---------|-------------|
| **Unit tests** | `make test` | Runs tests for all apps (with race detector and coverage). |
| **Lint** | `make lint` | Runs [golangci-lint](https://golangci-lint.run/) for all apps (`GOLANGCILINT_VERSION` is set in the workflow and in each app `Makefile`). |

Go is installed via `actions/setup-go` using the version from `apps/ddns/go.mod` (not `go.work`, which is not in the repository).



