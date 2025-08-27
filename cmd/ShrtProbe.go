/*
Copyright © 2025 Gi1gamesh666 <208263442@qq.com>
*/
package cmd

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/net/proxy"
)

type RequestConfig struct {
	httpClient   http.Client
	URL          string            // 请求URL（必需）
	Headers      map[string]string // 请求头（可选）
	Concurrency  int               // 最大并发数（默认5）
	RequestCount int64             // 总请求数（默认100）
	Timeout      time.Duration     // 请求超时时间（默认30秒）
}

type URLGenerator struct {
	baseURL    string
	charset    string
	pathLength int
	totalCount int
	current    int
}

var (
	successCount int64
	failureCount int64
)

func Proxy(proxyURL string) (*http.Client, error) {
	var transport http.Transport

	if proxyURL == "" {

		transport = http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		}
	} else {
		parsedURL, err := url.Parse(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("Invalid proxy URL: %v", err)
		}

		switch parsedURL.Scheme {
		case "http", "https":
			transport = http.Transport{
				Proxy: http.ProxyURL(parsedURL),
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
			}
		case "socks5":
			dialer, err := proxy.FromURL(parsedURL, proxy.Direct)
			if err != nil {
				return nil, fmt.Errorf("failed to create SOCKS5 proxy: %v", err)
			}
			contextDialer, ok := dialer.(proxy.ContextDialer)
			if !ok {
				return nil, fmt.Errorf("proxy dialer does not support ContextDialer")
			}
			transport = http.Transport{
				DialContext: contextDialer.DialContext,
			}
		}
	}

	return &http.Client{
		Transport: &transport,
		Timeout:   60 * time.Second,
	}, nil

}

func removeDuplicates(s string) string {
	seen := make(map[byte]bool)
	var result []byte

	for i := 0; i < len(s); i++ {
		if !seen[s[i]] {
			seen[s[i]] = true
			result = append(result, s[i])
		}
	}
	return string(result)
}

func GenerateRandomURL(baseURL, charset string, pathLength int) (string, error) {
	var sb strings.Builder

	// 确保baseURL以/结尾
	normalizedBaseURL := baseURL
	if normalizedBaseURL[len(normalizedBaseURL)-1] != '/' {
		normalizedBaseURL += "/"
	}
	sb.WriteString(normalizedBaseURL)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < pathLength; i++ {
		sb.WriteByte(charset[r.Intn(len(charset))])
	}

	return sb.String(), nil
}

func sendSingleRequest(config RequestConfig) (*http.Response, error) {

	client := httpClient

	req, err := http.NewRequestWithContext(
		context.Background(),
		"GET",
		config.URL,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("请求创建失败: %v", err)
	}

	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求发送失败: %v", err)
	}

	return resp, nil
}

func sendRequestsConcurrently(config RequestConfig) {
	var wg sync.WaitGroup
	taskChan := make(chan struct{}, config.Concurrency) // 改为发送空结构体信号

	// 重置计数器
	atomic.StoreInt64(&successCount, 0)
	atomic.StoreInt64(&failureCount, 0)

	// 启动消费者 goroutines
	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range taskChan {
				// 每次都生成新的随机URL
				randomURL, err := GenerateRandomURL(config.URL, charset, pathLength)
				if err != nil {
					atomic.AddInt64(&failureCount, 1)
					fmt.Printf("[-]URL生成失败: %v\n", err)
					continue
				}

				// 创建临时配置，使用新生成的URL
				tempConfig := config
				tempConfig.URL = randomURL

				resp, err := sendSingleRequest(tempConfig)
				if err != nil {
					atomic.AddInt64(&failureCount, 1)
					fmt.Printf("[-]请求[%s]失败: %v\n", randomURL, err)
					continue
				}

				// 只有3xx状态码才算成功
				if resp.StatusCode >= 300 && resp.StatusCode < 400 {
					atomic.AddInt64(&successCount, 1)
					location := resp.Header.Get("Location")
					if location != "" {
						fmt.Printf("[+]请求[%s]成功，状态码: %d，重定向到: %s\n", randomURL, resp.StatusCode, location)
					} else {
						fmt.Printf("[+]请求[%s]成功，状态码: %d\n", randomURL, resp.StatusCode)
					}
				} else {
					atomic.AddInt64(&failureCount, 1)
					fmt.Printf("[-]请求[%s]失败，状态码: %d\n", randomURL, resp.StatusCode)
				}

				if resp.Body != nil {
					resp.Body.Close()
				}
			}
		}()
	}

	// 生产者：发送任务信号到通道
	go func() {
		defer close(taskChan)
		for i := int64(0); i < config.RequestCount; i++ {
			taskChan <- struct{}{} // 发送空结构体作为信号
		}
	}()

	// 等待所有任务完成
	wg.Wait()

	// 输出统计结果
	fmt.Printf("\n统计结果:\n")
	fmt.Printf("成功请求: %d\n", atomic.LoadInt64(&successCount))
	fmt.Printf("失败请求: %d\n", atomic.LoadInt64(&failureCount))
	fmt.Printf("总请求数: %d\n", config.RequestCount)
}

// NewURLGenerator 创建新的URL生成器
func NewURLGenerator(baseURL, charset string, pathLength int) *URLGenerator {
	totalCount := 1
	for i := 0; i < pathLength; i++ {
		totalCount *= len(charset)
	}

	return &URLGenerator{
		baseURL:    baseURL,
		charset:    charset,
		pathLength: pathLength,
		totalCount: totalCount,
		current:    0,
	}
}

// HasNext 是否还有下一个URL
func (g *URLGenerator) HasNext() bool {
	return g.current < g.totalCount
}

// Next 生成下一个URL
func (g *URLGenerator) Next() string {
	if !g.HasNext() {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(g.baseURL)

	// 将当前数字转换为以charset长度为基数的表示
	num := g.current
	for i := 0; i < g.pathLength; i++ {
		sb.WriteByte(g.charset[num%len(g.charset)])
		num /= len(g.charset)
	}

	g.current++
	return sb.String()
}

// TotalCount 获取总数
func (g *URLGenerator) TotalCount() int {
	return g.totalCount
}

func sendRequestsConcurrentlyWithGenerator(config RequestConfig, generator *URLGenerator) {
	var wg sync.WaitGroup
	urlChan := make(chan string, config.Concurrency)

	// 启动工作协程
	for i := 0; i < config.Concurrency; i++ {
		go func() {
			for url := range urlChan {
				// 创建临时配置，修改URL
				tempConfig := config
				tempConfig.URL = url

				resp, err := sendSingleRequest(tempConfig)
				if err != nil {
					fmt.Printf("请求[%s]失败: %v\n", url, err)
					continue
				}
				fmt.Printf("请求[%s]成功，状态码: %d\n", url, resp.StatusCode)
				resp.Body.Close()
			}
			wg.Done()
		}()
	}

	// 发送任务
	wg.Add(1)
	go func() {
		defer close(urlChan)
		for generator.HasNext() {
			urlChan <- generator.Next()
		}
	}()

	wg.Wait()
}
