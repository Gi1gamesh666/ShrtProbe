# ShrtProbe

ShrtProbe 是一个用于测试 URL 短链接服务健壮性和安全性的工具。它能够生成随机或枚举的短链接路径，并通过并发 HTTP 请求对目标服务进行探测。

## 功能特性

- 生成随机或枚举的短链接路径
- 支持 HTTP/HTTPS/SOCKS5 代理
- 可自定义字符集、路径长度、并发数等参数
- 并发请求测试服务性能
- 伪装成正常浏览器请求
- 支持自定义请求头

## 安装

### 从源码编译

```bash
git clone https://github.com/yourusername/ShrtProbe.git
cd ShrtProbe
go build -o ShrtProbe
```


### 直接下载

从 [Releases](https://github.com/yourusername/ShrtProbe/releases) 页面下载预编译的二进制文件。

## 使用方法

### 基本用法

```bash
# 使用默认设置测试目标 URL
./ShrtProbe --url http://example.com/

# 使用自定义字符集和路径长度
./ShrtProbe --url http://example.com/ --charset abc123 --length 6

# 设置并发数和请求总数
./ShrtProbe --url http://example.com/ --concurrency 20 --count 1000

# 使用代理服务器
./ShrtProbe --url http://example.com/ --proxy http://proxy.example.com:8080
```


### 模式说明

ShrtProbe 支持两种探测模式：

1. **random（随机模式）**：随机生成指定长度的路径进行探测
2. **enumerate（枚举模式）**：按顺序枚举所有可能的路径组合进行探测

```bash
# 使用随机模式（默认）
./ShrtProbe --url http://example.com/ --mode random

# 使用枚举模式
./ShrtProbe --url http://example.com/ --mode enumerate
```


### 自定义请求头

```bash
# 添加自定义请求头
./ShrtProbe --url http://example.com/ --header "Referer: https://example.com/" --header "Cookie: sessionid=abc123"
```


## 命令行参数

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--url` | `-u` | 无（必需） | 目标 URL |
| `--charset` | `-s` | `abcdefghijklmnopqrstuvwxyz0123456789` | 字符集 |
| `--mode` | `-m` | `random` | 模式：`random` 或 `enumerate` |
| `--length` | `-l` | `5` | 路径长度 |
| `--count` | `-c` | `100` | 总请求数 |
| `--concurrency` | `-n` | `10` | 并发数 |
| `--proxy` | `-p` | 无 | 代理服务器地址 |
| `--header` | `-H` | 无 | 自定义请求头 |

## 输出示例

```
Using default charset: abcdefghijklmnopqrstuvwxyz0123456789
Proxy set successfully: 

Target URL:     https://example.com/
Mode:           random
Charset:        abcdefghijklmnopqrstuvwxyz0123456789
Path Length:    5
Request Count:  10
Concurrency:    15

[+]请求[https://example.com/abc12]成功，状态码: 302，重定向到: https://xxx.example.com/target
[-]请求[https://example.com/xyz99]失败，状态码: 404

统计结果:
成功请求: 1
失败请求: 9
总请求数: 10
Requests completed, total time: 308.169667ms
```


## 支持的代理类型

- HTTP 代理：`http://host:port`
- HTTPS 代理：`https://host:port`
- SOCKS5 代理：`socks5://host:port`

## 许可证

本项目采用 MIT 许可证。详情请见 [LICENSE](LICENSE) 文件。

## 免责声明

本工具仅供安全测试和教育目的使用。使用者应确保遵守相关法律法规，并仅在获得明确授权的情况下对目标系统进行测试。作者不对任何滥用本工具的行为负责。