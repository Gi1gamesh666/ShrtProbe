package probe

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// maxBodySample 为失败特征匹配读取响应体的上限（字节），避免大响应占满内存。
const maxBodySample = 512 * 1024

// Config 探测运行时参数（与 CLI 解耦，便于测试与复用）。
type Config struct {
	BaseURL      string
	Charset      string
	PathLength   int
	Headers      map[string]string
	Client       http.Client
	Timeout      time.Duration
	Concurrency  int
	RequestCount int64 // 仅 random：计划发起的任务数

	// FailBodyContains：响应体（不区分大小写）包含任一子串则视为「失败特征」命中。
	FailBodyContains []string
	// FailStatusCodes：状态码等于其中任一值则视为「失败特征」命中。
	FailStatusCodes []int
}

// HasFailRules 是否启用了失败特征判定（与仅看 3xx 的模式互斥）。
func (c *Config) HasFailRules() bool {
	return len(c.FailBodyContains) > 0 || len(c.FailStatusCodes) > 0
}

// needsBodySample 是否需要读取响应正文（仅配置了正文子串匹配时）。
func (c *Config) needsBodySample() bool {
	return len(c.FailBodyContains) > 0
}

// Outcome 单次探测的语义结果。
type Outcome struct {
	URL        string
	Status     int
	Location   string
	BodySample string // 用于匹配失败特征的前若干字节正文（UTF-8 文本场景）
	// FailReason 非空表示命中用户配置的「失败特征」（如 body 子串或状态码）。
	FailReason string
	Err        error // 网络 / HTTP 层错误
	GenErr     error // 仅 random：生成路径失败
}

// Hit 是否为一次「成功」探测：无传输错误，且未命中失败特征；未配置失败特征时沿用「HTTP 3xx」为成功。
func (c *Config) Hit(o Outcome) bool {
	if o.Err != nil || o.GenErr != nil {
		return false
	}
	if c.HasFailRules() {
		return o.FailReason == ""
	}
	return o.Status >= 300 && o.Status < 400
}

// Stats 聚合计数（由 Run* 返回，避免包级全局变量）。
type Stats struct {
	Success int64
	Failure int64
}

// Sequence 按字典序枚举短路径（base-N），与旧版 URLGenerator 行为一致。
type Sequence struct {
	baseURL    string
	charset    string
	pathLength int
	totalCount int
	current    int
}

// NewSequence 构造枚举器；baseURL 会规范化末尾斜杠。
func NewSequence(baseURL, charset string, pathLength int) *Sequence {
	normalized := strings.TrimSpace(baseURL)
	if normalized != "" && normalized[len(normalized)-1] != '/' {
		normalized += "/"
	}
	totalCount := 1
	for i := 0; i < pathLength; i++ {
		totalCount *= len(charset)
	}
	return &Sequence{
		baseURL:    normalized,
		charset:    charset,
		pathLength: pathLength,
		totalCount: totalCount,
		current:    0,
	}
}

func (s *Sequence) HasNext() bool {
	return s.current < s.totalCount
}

func (s *Sequence) Next() string {
	if !s.HasNext() {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(s.baseURL)
	num := s.current
	for i := 0; i < s.pathLength; i++ {
		sb.WriteByte(s.charset[num%len(s.charset)])
		num /= len(s.charset)
	}
	s.current++
	return sb.String()
}

func (s *Sequence) TotalCount() int {
	return s.totalCount
}

// RandomPath 在 baseURL 后拼接随机路径段。
func RandomPath(baseURL, charset string, pathLength int) (string, error) {
	if len(charset) == 0 {
		return "", fmt.Errorf("charset is empty")
	}
	if len(baseURL) == 0 {
		return "", fmt.Errorf("base URL is empty")
	}
	var sb strings.Builder
	normalized := baseURL
	if normalized[len(normalized)-1] != '/' {
		normalized += "/"
	}
	sb.WriteString(normalized)
	for i := 0; i < pathLength; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String(), nil
}

func (c *Config) failReason(status int, body string) string {
	for _, code := range c.FailStatusCodes {
		if status == code {
			return fmt.Sprintf("status=%d", code)
		}
	}
	if len(c.FailBodyContains) == 0 || body == "" {
		return ""
	}
	lower := strings.ToLower(body)
	for _, sub := range c.FailBodyContains {
		if sub == "" {
			continue
		}
		if strings.Contains(lower, strings.ToLower(sub)) {
			return fmt.Sprintf("body contains %q", sub)
		}
	}
	return ""
}

func doRequest(ctx context.Context, c *Config, url string) Outcome {
	reqCtx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		return Outcome{URL: url, Err: fmt.Errorf("build request: %w", err)}
	}
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return Outcome{URL: url, Err: err}
	}

	loc := ""
	if resp.Header != nil {
		loc = resp.Header.Get("Location")
	}

	var bodySample string
	if resp.Body != nil {
		if c.needsBodySample() {
			chunk, _ := io.ReadAll(io.LimitReader(resp.Body, maxBodySample))
			_, _ = io.Copy(io.Discard, resp.Body)
			bodySample = string(chunk)
		} else {
			_, _ = io.Copy(io.Discard, resp.Body)
		}
		_ = resp.Body.Close()
	}

	out := Outcome{
		URL:        url,
		Status:     resp.StatusCode,
		Location:   loc,
		BodySample: bodySample,
	}
	if c.HasFailRules() {
		out.FailReason = c.failReason(out.Status, out.BodySample)
	}
	return out
}

// RunRandom 并发随机路径探测；ctx 取消时尽快停止（如 Ctrl+C）。
func RunRandom(ctx context.Context, c Config, onEach func(Outcome)) Stats {
	var success, failure atomic.Int64

	taskCh := make(chan struct{}, c.Concurrency)
	var wg sync.WaitGroup

	for i := 0; i < c.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case _, ok := <-taskCh:
					if !ok {
						return
					}
					u, err := RandomPath(c.BaseURL, c.Charset, c.PathLength)
					if err != nil {
						failure.Add(1)
						if onEach != nil {
							onEach(Outcome{GenErr: err})
						}
						continue
					}
					out := doRequest(ctx, &c, u)
					if c.Hit(out) {
						success.Add(1)
					} else {
						failure.Add(1)
					}
					if onEach != nil {
						onEach(out)
					}
				}
			}
		}()
	}

	go func() {
		defer close(taskCh)
		for i := int64(0); i < c.RequestCount; i++ {
			select {
			case <-ctx.Done():
				return
			case taskCh <- struct{}{}:
			}
		}
	}()

	wg.Wait()
	return Stats{Success: success.Load(), Failure: failure.Load()}
}

// RunEnumerate 顺序枚举短路径并并发请求；成功判定与 RunRandom 相同（见 Config.Hit）。
func RunEnumerate(ctx context.Context, c Config, seq *Sequence, onEach func(Outcome)) Stats {
	if seq == nil {
		return Stats{}
	}

	urlCh := make(chan string, c.Concurrency)
	var success, failure atomic.Int64
	var wg sync.WaitGroup

	for i := 0; i < c.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for u := range urlCh {
				if ctx.Err() != nil {
					return
				}
				out := doRequest(ctx, &c, u)
				if c.Hit(out) {
					success.Add(1)
				} else {
					failure.Add(1)
				}
				if onEach != nil {
					onEach(out)
				}
			}
		}()
	}

	go func() {
		defer close(urlCh)
		for seq.HasNext() {
			select {
			case <-ctx.Done():
				return
			case urlCh <- seq.Next():
			}
		}
	}()

	wg.Wait()
	return Stats{Success: success.Load(), Failure: failure.Load()}
}
