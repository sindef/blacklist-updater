Fetches hosts-style blocklists from HTTP sources and saves them to disk. Validates content before writing and monitors sources for changes.

- Configure sources in `config.yaml`
- Run: `go run ./cmd/blacklistupdater -config config.yaml`
- Build binary: `go build ./cmd/blacklistupdater`
- Build container: `docker build -t blacklistupdater .`

Available as a container from dockerhub @ sindef/blacklistupdater

## Configuration

`config.yaml` structure:
- `sources`: List of HTTP sources with `url`, `filename`, and optional `output_format`
- `output_dir`: Directory to write files
- `interval_seconds`: Polling interval (0 = run once)
- `http_client.timeout_seconds`: HTTP request timeout

## Output Formats (optional)

- `hosts`: Standard hosts file (`0.0.0.0 domain.com`)
  - Strips regex patterns and wildcards
  - Validates domains after wildcard removal
  - Skips invalid domains (empty labels, leading/trailing dashes)

- `dnsmasq`: DNSmasq format (`address=/domain.com/0.0.0.0`)
  - Preserves wildcards (`*`)
  - Strips regex patterns

- `rfc1035`: DNS zone file format (`domain.com.	IN	A	0.0.0.0`)
  - Includes SOA record
  - Strips regex patterns and wildcards
  - Validates domains after wildcard removal
  - Requires trailing dots (FQDN format)