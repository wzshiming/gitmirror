# gitmirror

A read-only Git mirror server that proxies and caches Git repositories from upstream sources.

## Features

- **Read-only Mirror**: Serves Git repositories as read-only mirrors
- **On-demand Caching**: Automatically clones and caches repositories from upstream when first accessed
- **Smart HTTP Protocol**: Supports the Git Smart HTTP protocol for efficient transfers
- **Multiple Hosts**: Built-in support for GitHub, GitLab, Bitbucket, and Gitee

## Installation

```bash
go install github.com/wzshiming/gitmirror/cmd/gitmirror@latest
```

Or build from source:

```bash
git clone https://github.com/wzshiming/gitmirror.git
cd gitmirror
go build -o gitmirror ./cmd/gitmirror
```

## Usage

```bash
# Start the mirror server
gitmirror -addr :8080 -cache-dir ./cache

# Clone a repository through the mirror
git clone http://localhost:8080/github.com/owner/repo
```

### Command Line Options

| Option | Default | Description |
|--------|---------|-------------|
| `-addr` | `:8080` | HTTP server address |
| `-cache-dir` | `./cache` | Directory to store mirrored repositories |
| `-prefix` | `` | URL path prefix to strip |

## How It Works

1. When a client requests a repository (e.g., `git clone http://localhost:8080/github.com/owner/repo`), the server parses the URL to extract the repository path
2. The resolver maps the path to an upstream URL (e.g., `https://github.com/owner/repo.git`)
3. If the repository isn't cached locally, it's cloned as a bare mirror from the upstream
4. The server then serves the cached repository using the Git Smart HTTP protocol

## Supported Upstream Hosts

- `github.com`
- `gitlab.com`
- `bitbucket.org`
- `gitee.com`

## Architecture

```
┌─────────┐     ┌──────────────┐     ┌───────────────┐     ┌──────────┐
│  Client │────>│   Backend    │────>│    Mirror     │────>│ Upstream │
│ (git)   │<────│  (HTTP API)  │<────│   (cache)     │<────│  (git)   │
└─────────┘     └──────────────┘     └───────────────┘     └──────────┘
                       │
                       v
                ┌──────────────┐
                │   Resolver   │
                │ (URL parser) │
                └──────────────┘
```

## License

MIT License - see [LICENSE](LICENSE) for details.
