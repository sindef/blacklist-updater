Fetches hosts-style blocklists from HTTP sources and saves them to disk. Validates content before writing and monitors sources for changes.

• Configure sources in `config.yaml`
• Run: `go run ./cmd/dnsblacklist -config config.yaml`
• Build container: `docker build -t dnsblacklist .`