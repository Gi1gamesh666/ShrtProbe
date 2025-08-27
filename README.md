# ShrtProbe
[English](README-en.md) | [中文](README-zh.md)

ShrtProbe is a tool designed for testing the robustness and security of URL shortening services. It can generate random or enumerated short URL paths and probe target services through concurrent HTTP requests.

## Features

- Generate random or enumerated short URL paths
- Support HTTP/HTTPS/SOCKS5 proxies
- Customizable charset, path length, concurrency and other parameters
- Concurrent request testing for service performance
- Browser-like request to avoid detection
- Custom request headers support

## Installation

### Build from source

```bash
git clone https://github.com/yourusername/ShrtProbe.git
cd ShrtProbe
go build -o ShrtProbe
```


### Direct download

Download pre-compiled binaries from the [Releases](https://github.com/yourusername/ShrtProbe/releases) page.

## Usage

### Basic usage

```bash
# Test target URL with default settings
./ShrtProbe --url http://example.com/

# Use custom charset and path length
./ShrtProbe --url http://example.com/ --charset abc123 --length 6

# Set concurrency and request count
./ShrtProbe --url http://example.com/ --concurrency 20 --count 1000

# Use proxy server
./ShrtProbe --url http://example.com/ --proxy http://proxy.example.com:8080
```


### Mode explanation

ShrtProbe supports two probing modes:

1. **random mode**: Randomly generates paths of specified length for probing
2. **enumerate mode**: Sequentially enumerates all possible path combinations for probing

```bash
# Use random mode (default)
./ShrtProbe --url http://example.com/ --mode random

# Use enumerate mode
./ShrtProbe --url http://example.com/ --mode enumerate
```


### Custom request headers

```bash
# Add custom request headers
./ShrtProbe --url http://example.com/ --header "Referer: https://example.com/" --header "Cookie: sessionid=abc123"
```


## Command Line Arguments

| Parameter | Short | Default | Description |
|-----------|-------|---------|-------------|
| `--url` | `-u` | None (required) | Target URL |
| `--charset` | `-s` | `abcdefghijklmnopqrstuvwxyz0123456789` | Character set |
| `--mode` | `-m` | `random` | Mode: `random` or `enumerate` |
| `--length` | `-l` | `5` | Path length |
| `--count` | `-c` | `100` | Total request count |
| `--concurrency` | `-n` | `10` | Concurrency level |
| `--proxy` | `-p` | None | Proxy server address |
| `--header` | `-H` | None | Custom request headers |

## Output Example

```
Using default charset: abcdefghijklmnopqrstuvwxyz0123456789
Proxy set successfully: 

Target URL:     https://t.zsxq.com/
Mode:           random
Charset:        abcdefghijklmnopqrstuvwxyz0123456789
Path Length:    5
Request Count:  10
Concurrency:    15

[+]Request [https://example.com/abc12] succeeded, Status code: 302, Redirect to: https://xxx.example.com/target
[-]Request [https://example.com/xyz99] failed, Status code: 404

Statistics:
Successful requests: 1
Failed requests: 9
Total requests: 10
Requests completed, total time: 308.169667ms
```


## Supported Proxy Types

- HTTP proxy: `http://host:port`
- HTTPS proxy: `https://host:port`
- SOCKS5 proxy: `socks5://host:port`

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Disclaimer

This tool is for security testing and educational purposes only. Users should ensure compliance with relevant laws and regulations and only test target systems with explicit authorization. The author is not responsible for any misuse of this tool.
