/*
Copyright © 2025 Gi1gamesh666 <208263442@qq.com>
*/
package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"time"
)

var (
	httpClient   http.Client = http.Client{}
	proxyAddr    string
	requestCount int64
	charset      string
	mode         string
	concurrency  int
	addColor     = color.New(color.FgGreen).Add(color.Bold).PrintfFunc()
	removeColor  = color.New(color.FgRed).Add(color.Bold).PrintfFunc()
	errorColor   = color.New(color.FgRed).Add(color.Bold).PrintfFunc()
	config       = RequestConfig{
		Method:       "GET",
		URL:          "",
		Timeout:      10 * time.Second,
		RequestCount: 100,
		Concurrency:  10,
	}
)

var rootCmd = &cobra.Command{
	Use:   "ShrtProbe",
	Short: "A robustness and security testing tool for URL shortening services.",

	Run: func(cmd *cobra.Command, args []string) {

		if config.URL == "" {
			errorColor("请设置目标URL")
			os.Exit(1)
		}

		if charset == "" {
			addColor("请设置字符集: (默认字符集为: abcdefghijklmnopqrstuvwxyz0123456789)")

		} else {
			addColor("已设置字符集: %s\n", charset)
		}

		if mode == "enumerate" {
			config.URL, _ = GenerateEnumerateURL(config.URL, charset, 5, int(requestCount))
		} else if mode == "random" {
			config.URL, _ = GenerateRandomURL(config.URL, charset, 5)
		} else {
			errorColor("模式设置异常: %s\n", mode)
		}

		startTime := time.Now()
		duration := time.Since(startTime)
		fmt.Printf("请求完成，总耗时: %v\n", duration)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "查看应用版本",
	Aliases: []string{"v"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Version: 1.0.0")
	},
}

var proxycmd = &cobra.Command{
	Use:   "proxy",
	Short: "设置代理",
	Run: func(cmd *cobra.Command, args []string) {
		proxy, err := Proxy(proxyAddr)
		if err != nil {
			removeColor("代理设置异常: %w", err)
		} else {
			httpClient = *proxy
			addColor("代理已设置成功: %s\n", proxyAddr)
		}
	},
}

var requestCountCmd = &cobra.Command{
	Use:   "requestCount",
	Short: "设置请求数",
	Run: func(cmd *cobra.Command, args []string) {
		addColor("已设置请求数: %d\n", requestCount)
	},
}

var concurrencyCmd = &cobra.Command{
	Use:   "concurrency",
	Short: "设置并发数",
	Run: func(cmd *cobra.Command, args []string) {
		addColor("已设置并发数: %d\n", concurrency)
	},
}

func init() {

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().StringVarP(&config.URL, "url", "u", "", "请求的目标URL（必需）")
	proxycmd.Flags().StringVarP(&proxyAddr, "proxy", "p", "", "设置代理服务器地址（格式：http://host:port）")
	requestCountCmd.Flags().Int64VarP(&requestCount, "count", "c", 100, "设置总请求数（默认100）")
	concurrencyCmd.Flags().IntVarP(&concurrency, "concurrency", "n", 10, "设置并发数（默认10）")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(proxycmd)
	rootCmd.AddCommand(requestCountCmd)
	rootCmd.AddCommand(concurrencyCmd)

}
