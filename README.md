# ShrtProbe

English | [中文](README-zh.md)

ShrtProbe probes URL shortening services by generating random or enumerated short paths and issuing concurrent HTTP requests. Core logic lives in [`internal/probe`](internal/probe); the CLI is a thin [`cmd`](cmd) layer.

## Features

- **Modes**: `random` (random paths) and `enumerate` (exhaustive path sequence in charset order)
- **Proxies**: HTTP, HTTPS, `socks5`, `socks5h`
- **Concurrency**: worker pool with connection reuse and configurable per-request timeout
- **Success criteria** (see below): default **HTTP 3xx**, or **custom failure rules** (body substring / status code) where success means *not* matching any failure rule
- **Graceful stop**: `Ctrl+C` cancels in-flight work via `context`
- Default browser-like headers (built-in; not configurable via flags today)

## Installation

### Build from source

Requires [Go](https://go.dev/dl/) 1.21+ (see `go.mod`).

```bash
git clone https://github.com/Gi1gamesh666/ShrtProbe.git
cd ShrtProbe
go build -o ShrtProbe .
```

### Prebuilt binaries

Download from [Releases](https://github.com/Gi1gamesh666/ShrtProbe/releases) if available.

## Usage

### Basic

```bash
./ShrtProbe --url http://example.com/

./ShrtProbe --url http://example.com/ --charset abc123 --length 6

./ShrtProbe --url http://example.com/ --concurrency 20 --count 1000

./ShrtProbe --url http://example.com/ --proxy http://proxy.example.com:8080
```

### Modes

| Mode | Behavior |
|------|----------|
| `random` (default) | Sends `--count` probes; each uses a random path of length `--length` |
| `enumerate` | Walks every path combination in order; total = `len(charset) ^ length` |

```bash
./ShrtProbe --url http://example.com/ --mode random
./ShrtProbe --url http://example.com/ --mode enumerate
```

### Success criteria

**Default (no failure rules)**  
A probe is **successful** when the HTTP response status is **3xx** (typical short-link redirect).

**With `--fail-body` / `--fail-status`**  
You define what counts as **failure**. A probe is **successful** when the request completes without error and **does not** match any failure rule:

- **`--fail-body`** (repeatable): failure if the response body (first 512 KiB, case-insensitive) contains the substring.
- **`--fail-status`** (repeatable): failure if the status code equals the given value.

Status rules are checked first, then body rules. If only `--fail-status` is set, the body is not read.

```bash
./ShrtProbe --url http://example.com/ \
  --fail-body "not found" --fail-body "页面不存在" \
  --fail-status 404 --fail-status 410
```

### Other commands

```bash
./ShrtProbe version
./ShrtProbe completion bash   # zsh | fish | powershell
```

## Command-line flags

| Flag | Shorthand | Default | Description |
|------|-----------|---------|-------------|
| `--url` | `-u` | *(required)* | Base URL of the short-link service |
| `--charset` | `-s` | `a-zA-Z0-9` | Character set for path segments |
| `--mode` | `-m` | `random` | `random` or `enumerate` |
| `--length` | `-l` | `5` | Path length (number of characters) |
| `--count` | `-c` | `10` | Number of requests (**random** mode only) |
| `--concurrency` | `-n` | `15` | Concurrent workers |
| `--timeout` | `-t` | `10s` | Per-request timeout |
| `--proxy` | `-p` | *(empty)* | Proxy URL, e.g. `http://host:port`, `socks5://host:port` |
| `--fail-body` | | *(none)* | Failure if body contains substring (repeatable) |
| `--fail-status` | | *(none)* | Failure if status equals value (repeatable) |

Run `./ShrtProbe --help` for the full help text.

## Output example

```
Using default charset: abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789

Target URL:     https://example.com/
Mode:           random
Charset:        abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789
Path Length:    5
Request Count:  10
Concurrency:    15
成功判定:       HTTP 3xx 重定向

[+]请求[https://example.com/abc12]成功，状态码: 302，重定向到: https://target.example/page
[-]请求[https://example.com/xyz99]失败，状态码: 404（成功条件: HTTP 3xx）

统计结果:
成功请求: 1
失败请求: 9
总请求数: 10
Requests completed, total time: 308ms
```

## Proxy URL formats

- HTTP: `http://host:port`
- HTTPS: `https://host:port`
- SOCKS5: `socks5://host:port` or `socks5h://host:port` (remote DNS)

## License

See [LICENSE](LICENSE).

## Disclaimer

For authorized security testing and education only. Do not use against systems you do not own or lack permission to test. Authors are not responsible for misuse.
