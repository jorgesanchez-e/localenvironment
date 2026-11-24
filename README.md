# localenvironment

Monorepo for building a **local lab**. Each lab component is developed or configured as an independent project inside this same repository.

## Components

| # | Component | Location | Status |
|---|-----------|----------|--------|
| 1 | **NanoBSD** image that will act as the lab's router/firewall | — | Pending |
| 2 | TLS certificate issuance with **ACME / Let's Encrypt** (DNS-01 challenge over Route53) | [apps/acme](apps/acme) | In progress |
| 3 | **No-IP**-style client in Go that detects the router's public IPv4 and IPv6 and updates DNS (Route53 for now) | — | Pending |

### 1. NanoBSD router/firewall (pending)

Configuration to build a NanoBSD image that works as the lab's router and firewall. Not started yet.

### 2. TLS certificates with ACME (in progress)

Go application ([apps/acme](apps/acme)) that requests TLS certificates from Let's Encrypt using the [lego](https://github.com/go-acme/lego) library. It validates domains with the **DNS-01** challenge and uses **AWS Route53** as the DNS server.

Implemented so far:

- **Domain** ([internal/domain/certs](apps/acme/internal/domain/certs)): the `Cert` entity (domain, certificate, private key, issuer certificate and CSR), the logic that decides whether a certificate needs renewal (`NeedsRenewal`), and the `Driver` (issue/revoke) and `Repository` (save/load) interfaces.
- **ACME service** ([internal/infra/acme](apps/acme/internal/infra/acme)): picks the implementations based on the configuration and only sends empty certificates, or ones that need renewal, to be issued.
- **AWS + lego driver** ([internal/infra/acme/aws](apps/acme/internal/infra/acme/aws)):
  - registers the ACME account with an email and a PEM private key;
  - supports several AWS accounts, each with its own domains and Route53 `zone-id`;
  - issues certificates in parallel (EC256 key, chain included);
  - validates the configuration with `go-playground/validator`.
  - It currently points to the Let's Encrypt *staging* environment. Revocation (`Revoke`) is not implemented yet.
- **Persistence** ([internal/infra/store](apps/acme/internal/infra/store)): stores certificates in **SQLite** (`modernc.org/sqlite`, no CGO). The `certs` table is created automatically on startup.
- **Entry point** ([cmd/main.go](apps/acme/cmd/main.go)): loads the configuration, issues the certificates and saves them. For now it uses test domains hardcoded in the source.

#### Configuration

The app reads a `config.yaml` file from the first of these paths where it finds one: `/usr/local/etc/localenvironment/`, `/etc/localenvironment/`, `$HOME/.localenvironment/` or the current directory.

```yaml
acme:
  email: user@example.com
  primary_key: |
    -----BEGIN EC PRIVATE KEY-----
    ...
    -----END EC PRIVATE KEY-----
  aws:
    - account: lab
      access-key: AKIA...
      secret-key: ...
      region: us-east-1
      domains:
        - name: test-a.example.dev
          zone-id: Z0123456789ABC
  store:
    type: sqlite
    sqlite:
      file: ./certs.db
```

### 3. Dynamic DNS client (pending)

No-IP-style Go client that detects the router's public IPv4 and IPv6 and updates the records on the DNS server (Route53 at first). Not started yet.

## Shared libraries

- [pkg/config](pkg/config): YAML configuration loading with [viper](https://github.com/spf13/viper). Used by the projects in the monorepo.

## Layout

```
.
├── apps/
│   └── acme/          # TLS certificate issuance (ACME + Route53)
├── pkg/
│   └── config/        # shared configuration library
├── go.work            # Go workspace that ties the modules together
└── Makefile           # runs build/test/lint across all projects
```

Each project is an independent Go module with its own `Makefile` and scripts under `scripts/`. The root `go.work` ties them together for local development.

## Usage

Requirements: Go, git, make (GNU or BSD) and, for linting, curl.

From the root:

```sh
make build        # builds all apps into build/
make test         # unit tests for apps and packages
make test-report  # shows the coverage report
make lint         # golangci-lint
make clean
```

From a project, for example `apps/acme`:

```sh
make run          # builds and runs
make docker-build # builds the Docker image
```

Useful variables: `GOOS`, `GOARCH` and `APPVERSION` (defaults to `0.1`).
