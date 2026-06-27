# Vault

[![vault](https://github.com/Gaucho-Racing/Vault/actions/workflows/vault.yml/badge.svg)](https://github.com/Gaucho-Racing/Vault/actions/workflows/vault.yml)
[![Release](https://img.shields.io/github/v/release/Gaucho-Racing/Vault?style=flat-square)](https://github.com/Gaucho-Racing/Vault/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Vault is Gaucho Racing's internal secrets manager for shared credentials, application secrets, and workflow automation secrets.
It provides a Sentinel-backed web interface for storing account credentials, TOTP seeds, notes, API keys, and app-scoped secrets with group-based access controls.

Vault also powers GitHub Actions secret delivery through OIDC.
Repositories can request explicit app-secret selectors, while Vault centrally evaluates repository, ref, and selector rules before exporting secrets into the workflow environment.

Production: [vault.gauchoracing.com](https://vault.gauchoracing.com)

## Features

- Sentinel SSO and group-based access for accounts and app-secret applications.
- Encrypted account secrets for passwords, TOTP seeds, API keys, URLs, notes, and custom secret types.
- App secrets referenced by selectors such as `mapache-prod.sentinel_client_id`.
- GitHub Actions OIDC rules for exporting selected app secrets to trusted workflows.
- Audit logs for account and secret views, with duplicate view events debounced.
- Multi-architecture server and web images published to GitHub Container Registry.

## Getting Started

Run the local development stack with Docker Compose:

```bash
docker compose up --build
```

The development proxy serves Vault at:

```text
http://localhost:10310
```

The API is available under:

```text
http://localhost:10310/api
```

The compose stack starts:

- `vault`: Go API server
- `web`: React/Vite frontend
- `db`: PostgreSQL
- `kerbecs`: local reverse proxy

## Development

Run backend checks:

```bash
cd vault
go test ./...
```

Run frontend checks:

```bash
cd web
npm install
npm run lint
npm run build
```

## GitHub Actions Secrets

Vault exposes a GitHub Actions OIDC export endpoint for app secrets.
Rules are managed from the Vault settings page and can be created by users with `sentinel:all`, `Admins`, or `DevopsMembers` access.

Use the dedicated action repository to pull secrets in workflows:

- [Gaucho-Racing/vault-pull-secrets](https://github.com/Gaucho-Racing/vault-pull-secrets)

Example:

```yaml
permissions:
  id-token: write
  contents: read

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: Gaucho-Racing/vault-pull-secrets@v1
        with:
          secrets: pypi.publish_token
```

## Release

Create a new Vault release from an up-to-date `main` branch:

```bash
scripts/release.sh 1.4.0
```

The release workflow publishes versioned `vault-server` and `vault-web` images, then opens an infrastructure PR to deploy the new image tags.

## Related Projects

- [Sentinel](https://github.com/Gaucho-Racing/Sentinel): authentication and access management
- [Vault Pull Secrets](https://github.com/Gaucho-Racing/vault-pull-secrets): GitHub Action for exporting Vault app secrets

## Contributing

1. Fork the project.
2. Create your feature branch (`git checkout -b gh-username/my-amazing-feature`).
3. Commit your changes (`git commit -m 'Add my amazing feature'`).
4. Push to the branch (`git push origin gh-username/my-amazing-feature`).
5. Open a pull request.
