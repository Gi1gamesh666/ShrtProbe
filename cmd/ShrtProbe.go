package cmd

import (
	"fmt"
	"golang.org/x/net/proxy"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
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
				return nil, fmt.Errorf("Failed to create SOCKS5 proxy: %v", err)
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
	defaultcharset := "abcdefghijklmnopqrstuvwxyz0123456789"
	var sb strings.Builder
	sb.WriteString(baseURL)
	if charset == "" {
		charset = defaultcharset
	} else {
		charset = removeDuplicates(defaultcharset + charset)
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < pathlenth; i++ {
		sb.WriteByte(charset[r.Intn(len(charset))])
	}

	return sb.String(), nil
}

func GenerateEnumerateURL(baseURL, charset string, pathlenth int, counts int) (string, error) {
	defaultcharset := "abcdefghijklmnopqrstuvwxyz0123456789"
	var sb strings.Builder
	var urls []string
	sb.WriteString(baseURL)
	if charset == "" {
		charset = defaultcharset
	} else {
		charset = removeDuplicates(defaultcharset + charset)
	}

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
