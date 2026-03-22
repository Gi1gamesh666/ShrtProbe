package probe

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/proxy"
)

func dialer() *net.Dialer {
	return &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
}

// NewBaseTransport 适合高并发的默认传输层（连接复用、空闲连接上限）。
func NewBaseTransport() *http.Transport {
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer().DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// NewHTTPClient 根据 proxyURL 构建客户端；proxyURL 为空时使用系统代理环境变量与默认传输层。
func NewHTTPClient(proxyURL string, timeout time.Duration) (*http.Client, error) {
	if proxyURL == "" {
		return &http.Client{
			Transport: NewBaseTransport(),
			Timeout:   timeout,
		}, nil
	}

	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}

	var transport *http.Transport

	switch parsedURL.Scheme {
	case "http", "https":
		transport = NewBaseTransport()
		transport.Proxy = http.ProxyURL(parsedURL)
	case "socks5", "socks5h":
		socksDialer, err := proxy.FromURL(parsedURL, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("failed to create SOCKS5 proxy: %w", err)
		}
		contextDialer, ok := socksDialer.(proxy.ContextDialer)
		if !ok {
			return nil, fmt.Errorf("proxy dialer does not support ContextDialer")
		}
		transport = NewBaseTransport()
		transport.Proxy = nil
		transport.DialContext = contextDialer.DialContext
	default:
		return nil, fmt.Errorf("unsupported proxy scheme %q (use http, https, socks5)", parsedURL.Scheme)
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}, nil
}
