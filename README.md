Fetches hosts-style blocklists from HTTP sources and saves them to disk. Validates content before writing and monitors sources for changes.

- Configure sources in `config.yaml`
- Run: `go run ./cmd/blacklistupdater -config config.yaml`
- Build binary: `go build ./cmd/blacklistupdater`
- Build container: `docker build -t blacklistupdater .`

Available as a container from dockerhub @ sindef/blacklistupdater