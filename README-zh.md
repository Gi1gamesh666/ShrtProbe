# ShrtProbe

[English](README.md) | 中文

ShrtProbe 用于对 URL 短链接服务做探测：随机或按字典序枚举短路径，并发起并发 HTTP 请求。核心逻辑在 [`internal/probe`](internal/probe)，命令行入口在 [`cmd`](cmd)。

## 功能特性

- **模式**：`random`（随机路径）、`enumerate`（按字符集顺序枚举全部路径组合）
- **代理**：HTTP、HTTPS、`socks5`、`socks5h`
- **并发与超时**：工作池、连接复用、单次请求超时（`--timeout`）
- **成功判定**（见下文）：默认「**HTTP 3xx**」；可改为「**失败特征**」——你定义「什么叫失败」，**未命中任一失败特征即成功**
- **中断**：`Ctrl+C` 通过 `context` 取消，便于停止
- 内置类浏览器请求头（当前版本**无** `--header` 类自定义头参数）

## 安装

### 从源码编译

```bash
git clone https://github.com/Gi1gamesh666/ShrtProbe.git
cd ShrtProbe
go build -o ShrtProbe .
```

### 预编译二进制

若仓库提供 [Releases](https://github.com/Gi1gamesh666/ShrtProbe/releases)，可从中下载。

## 使用方法

### 基本示例

```bash
./ShrtProbe --url http://example.com/

./ShrtProbe --url http://example.com/ --charset abc123 --length 6

./ShrtProbe --url http://example.com/ --concurrency 20 --count 1000

./ShrtProbe --url http://example.com/ --proxy http://proxy.example.com:8080
```

### 模式说明

| 模式 | 说明 |
|------|------|
| `random`（默认） | 共发起 `--count` 次请求，每次路径为 `--length` 长度的随机串 |
| `enumerate` | 按序枚举所有 `len(charset)^length` 条路径 |

```bash
./ShrtProbe --url http://example.com/ --mode random
./ShrtProbe --url http://example.com/ --mode enumerate
```

### 成功判定

**未配置失败特征时**  
与常见短链行为一致：**HTTP 状态码为 3xx** 视为一次成功探测。

**配置了 `--fail-body` 和/或 `--fail-status` 时**  
由你定义「失败特征」；**成功 = 请求无错误，且未命中任一失败特征**（不再强制要求 3xx）：

- **`--fail-body`**（可重复）：响应正文（前 512 KiB，**不区分大小写**）包含该子串则判失败。
- **`--fail-status`**（可重复）：HTTP 状态码等于该值则判失败。

判定顺序：先匹配状态码列表，再匹配正文子串。若**只**配置了 `--fail-status`，则**不读取**正文，节省流量。

```bash
./ShrtProbe --url http://example.com/ \
  --fail-body "not found" --fail-body "页面不存在" \
  --fail-status 404 --fail-status 410
```

### 子命令

```bash
./ShrtProbe version
./ShrtProbe completion bash   # 亦支持 zsh、fish、powershell
```

## 命令行参数

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--url` | `-u` | 无（必填） | 短链服务基础 URL |
| `--charset` | `-s` | `a-zA-Z0-9` | 路径字符集 |
| `--mode` | `-m` | `random` | `random` 或 `enumerate` |
| `--length` | `-l` | `5` | 路径长度 |
| `--count` | `-c` | `10` | 请求总数（仅 **random**） |
| `--concurrency` | `-n` | `15` | 并发 worker 数 |
| `--timeout` | `-t` | `10s` | 单次请求超时 |
| `--proxy` | `-p` | 无 | 代理，如 `http://host:port`、`socks5://host:port` |
| `--fail-body` | | 无 | 正文包含子串则失败（可重复） |
| `--fail-status` | | 无 | 状态码等于该值则失败（可重复） |

完整说明请执行：`./ShrtProbe --help`。

## 输出示例

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

## 支持的代理格式

- HTTP：`http://host:port`
- HTTPS：`https://host:port`
- SOCKS5：`socks5://host:port` 或 `socks5h://host:port`（远端解析）

## 许可证

见 [LICENSE](LICENSE)。

## 免责声明

本工具仅供在**授权范围内**的安全测试与学习使用。请勿对未获授权的系统使用。作者不对滥用行为负责。
