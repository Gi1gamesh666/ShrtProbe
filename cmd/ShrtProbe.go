/*
Copyright © 2025 Gi1gamesh666 <208263442@qq.com>
*/
package cmd

import (
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/net/proxy"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type RequestConfig struct {
	httpClient   http.Client
	URL          string            // 请求URL（必需）
	Headers      map[string]string // 请求头（可选）
	Concurrency  int               // 最大并发数（默认5）
	RequestCount int64             // 总请求数（默认100）
	Timeout      time.Duration     // 请求超时时间（默认30秒）
}

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

func GenerateRandomURL(baseURL, charset string, pathlenth int) (string, error) {
	var sb strings.Builder
	sb.WriteString(baseURL)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < pathlenth; i++ {
		sb.WriteByte(charset[r.Intn(len(charset))])
	}

	return sb.String(), nil
}

func GenerateEnumerateURL(baseURL, charset string, pathlenth int, counts int) (string, error) {
	var sb strings.Builder
	var urls []string
	sb.WriteString(baseURL)

	for i := 0; i < counts; i++ {
		for i := 0; i < pathlenth; i++ {
			sb.WriteByte(charset[i%len(charset)])
			urls = append(urls, sb.String())
		}
	}

	tmpFile, err := os.CreateTemp("", "enumerate.txt")
	if err != nil {
		return "", fmt.Errorf("Failed to create temporary file: %v", err)
	}

	for _, u := range urls {
		_, err := tmpFile.WriteString(u + "\n")
		if err != nil {
			err := tmpFile.Close()
			if err != nil {
				return "", fmt.Errorf("Failed to close temporary file: %v", err)
			}
			err = os.Remove(tmpFile.Name())
			if err != nil {
				return "", fmt.Errorf("Failed to remove temporary file: %v", err)
			}
			return "", fmt.Errorf("Failed to write to temporary file: %v", err)
		}
	}

	err = tmpFile.Close()
	if err != nil {
		err := os.Remove(tmpFile.Name())
		if err != nil {
			return "", fmt.Errorf("Failed to remove temporary file: %v", err)
		}
		return "", fmt.Errorf("Failed to close temporary file: %v", err)
	}

	return tmpFile.Name(), nil
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
	taskChan := make(chan string, config.Concurrency)

	for i := int64(0); i < config.RequestCount; i++ {
		wg.Add(1)
		taskChan <- config.URL
	}

	go func() {
		wg.Wait()
		close(taskChan)
	}()

	for url := range taskChan {
		go func(u string) {
			defer wg.Done()

			resp, err := sendSingleRequest(config) // 假设sendSingleRequest支持传入URL
			if err != nil {
				fmt.Printf("请求[%s]失败: %v\n", config.URL, err) // 直接打印URL
				return
			}
			fmt.Printf("请求[%s]成功，状态码: %d\n", config.URL, resp.StatusCode) // 直接打印URL
		}(url)
	}
}
