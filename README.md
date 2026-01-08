# gitmirror

A read-only Git mirror server that proxies Git repositories from upstream sources with real-time request forwarding.

## Features

- **Real-time Proxy**: Proxies git requests directly to upstream without pre-cloning
- **Request Parsing**: Parses git-upload-pack requests to understand client needs
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
3. The server proxies the git-upload-pack request to the upstream in real-time
4. The response is streamed directly back to the client

## Supported Upstream Hosts

- `github.com`
- `gitlab.com`
- `bitbucket.org`
- `gitee.com`

## Architecture

```
┌─────────┐     ┌──────────────┐     ┌───────────────┐     ┌──────────┐
│  Client │────>│   Backend    │────>│    Proxy      │────>│ Upstream │
│ (git)   │<────│  (HTTP API)  │<────│  (forward)    │<────│  (git)   │
└─────────┘     └──────────────┘     └───────────────┘     └──────────┘
                       │
                       v
                ┌──────────────┐
                │   pktline    │
                │  (parser)    │
                └──────────────┘
```

### Key Components

- **`pkg/backend/`** - Git Smart HTTP protocol handler
- **`pkg/proxy/`** - Upstream request proxying
- **`pkg/pktline/`** - Git pkt-line format parser
- **`pkg/upstream/`** - URL resolution for upstream hosts
- **`pkg/mirror/`** - Repository management

## License

MIT License - see [LICENSE](LICENSE) for details.
