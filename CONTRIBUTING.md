# Contributing to Kora

Kora is a config-driven application engine. We welcome contributions of all kinds — configs, docs, bug fixes, features, and ideas.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/kora.git`
3. Read [SETUP.md](docs/SETUP.md) to get a working development environment
4. Create a branch: `git checkout -b feature/your-feature`

## Development

```bash
# Backend
go run . serve --port 8000          # Run the server
go test ./...                       # Run tests
go build -o kora .                  # Build binary

# Frontend
cd ui && bun install                # Install dependencies
cd ui && bun run dev                # Dev server (proxies API to :8000)
cd ui && bun run build              # Production build
```

## What to Work On

- **Sample apps** — new config-driven applications under `config/`
- **Field types** — new field renderers in the React SPA
- **Expression functions** — extend `ui/src/lib/expression-eval.ts`
- **Backend drivers** — PostgreSQL support, S3 storage
- **Documentation** — improve or translate docs
- **Bug fixes** — check open issues

## Pull Requests

- Keep PRs focused on one thing
- Update docs if your change adds or changes behavior
- Run `go build -o /dev/null .` and `cd ui && bun run build` before submitting
- Describe what you changed and why

## License

Kora is licensed under the GNU Affero General Public License v3.0. By contributing, you agree that your contributions will be licensed under the same terms.
