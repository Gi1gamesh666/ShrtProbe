/*
Copyright © 2025 Gi1gamesh666 <208263442@qq.com>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/Gi1gamesh666/ShrtProbe/internal/probe"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	proxyAddr      string
	failBody       []string
	failStatus     []int
	defaultCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charset        string
	mode           string
	pathLength     int
	addColor       = color.New(color.FgGreen).Add(color.Bold).PrintfFunc()
	removeColor    = color.New(color.FgRed).Add(color.Bold).PrintfFunc()
	errorColor     = color.New(color.FgRed).Add(color.Bold).PrintfFunc()
	config         = probe.Config{
		Timeout:      10 * time.Second,
		RequestCount: 10,
		Concurrency:  15,
		Headers: map[string]string{
			"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			"Accept-Language":           "en-US,en;q=0.5",
			"Accept-Encoding":           "gzip, deflate",
			"Connection":                "keep-alive",
			"Upgrade-Insecure-Requests": "1",
			"Sec-Fetch-Dest":            "document",
			"Sec-Fetch-Mode":            "navigate",
			"Sec-Fetch-Site":            "none",
			"Cache-Control":             "max-age=0",
		},
	}
)

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

func printOutcome(cfg *probe.Config) func(probe.Outcome) {
	return func(o probe.Outcome) {
		if o.GenErr != nil {
			removeColor("[-]URL生成失败: %v\n", o.GenErr)
			return
		}
		if o.Err != nil {
			removeColor("[-]请求[%s]失败: %v\n", o.URL, o.Err)
			return
		}
		if cfg.Hit(o) {
			if o.Location != "" {
				addColor("[+]请求[%s]成功，状态码: %d，重定向到: %s\n", o.URL, o.Status, o.Location)
			} else {
				addColor("[+]请求[%s]成功，状态码: %d\n", o.URL, o.Status)
			}
			return
		}
		if o.FailReason != "" {
			removeColor("[-]请求[%s]失败，命中失败特征: %s (状态码: %d)\n", o.URL, o.FailReason, o.Status)
			return
		}
		removeColor("[-]请求[%s]失败，状态码: %d（成功条件: HTTP 3xx）\n", o.URL, o.Status)
	}
}

var rootCmd = &cobra.Command{
	Use:   "ShrtProbe",
	Short: "A robustness and security testing tool for URL shortening services.",
	Long: `ShrtProbe is a tool designed for testing the robustness and security of URL shortening services.

Key features:
  - Generate random or enumerated short URL paths
  - Send concurrent HTTP requests to test service
  - Support HTTP/HTTPS/SOCKS5 proxies
  - Customizable charset, path length, concurrency and other parameters

Usage examples:
  # Test target URL with default settings
  ShrtProbe --url http://example.com/

  # Use custom charset and path length
  ShrtProbe --url http://example.com/ --charset abc123 --length 6

  # Set concurrency and request count
  ShrtProbe --url http://example.com/ --concurrency 20 --count 1000

  # Use proxy server
  ShrtProbe --url http://example.com/ --proxy http://proxy.example.com:8080

  # Define failure by body substring / status (success = not matching any failure rule)
  ShrtProbe --url http://example.com/ --fail-body "not found" --fail-status 404`,

	Run: func(cmd *cobra.Command, args []string) {
		if config.BaseURL == "" {
			errorColor("Error: Target URL is required\n")
			os.Exit(1)
		}

		hc, err := probe.NewHTTPClient(proxyAddr, config.Timeout)
		if err != nil {
			errorColor("HTTP client setup error: %v\n", err)
			os.Exit(1)
		}
		config.Client = *hc
		if proxyAddr != "" {
			addColor("Proxy set successfully: %s\n", proxyAddr)
		}

		if charset == "" {
			addColor("Using default charset: %s\n", defaultCharset)
			charset = defaultCharset
		} else {
			charset = removeDuplicates(charset)
			addColor("Charset set: %s\n", charset)
		}

		if mode != "enumerate" && mode != "random" {
			removeColor("Error: Unknown mode. Use 'enumerate' or 'random'\n")
			os.Exit(1)
		}

		if pathLength <= 0 {
			removeColor("Error: Path length must be greater than 0\n")
			os.Exit(1)
		}

		fmt.Println("")
		addColor("Target URL:     %s\n", config.BaseURL)
		addColor("Mode:           %s\n", mode)
		addColor("Charset:        %s\n", charset)
		addColor("Path Length:    %d\n", pathLength)
		addColor("Request Count:  %d\n", config.RequestCount)
		addColor("Concurrency:    %d\n", config.Concurrency)
		if len(failBody) > 0 || len(failStatus) > 0 {
			addColor("成功判定:       未命中任一失败特征（与状态码 3xx 无关）\n")
			if len(failBody) > 0 {
				addColor("失败特征(正文): %v\n", failBody)
			}
			if len(failStatus) > 0 {
				addColor("失败特征(状态): %v\n", failStatus)
			}
		} else {
			addColor("成功判定:       HTTP 3xx 重定向\n")
		}

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()

		runCfg := config
		runCfg.Charset = charset
		runCfg.PathLength = pathLength
		runCfg.FailBodyContains = append([]string(nil), failBody...)
		runCfg.FailStatusCodes = append([]int(nil), failStatus...)

		emit := printOutcome(&runCfg)

		if mode == "random" {
			if runCfg.BaseURL[len(runCfg.BaseURL)-1] != '/' {
				runCfg.BaseURL += "/"
			}

			startTime := time.Now()
			stats := probe.RunRandom(ctx, runCfg, emit)
			duration := time.Since(startTime)
			fmt.Printf("\n统计结果:\n")
			fmt.Printf("成功请求: %d\n", stats.Success)
			fmt.Printf("失败请求: %d\n", stats.Failure)
			fmt.Printf("总请求数: %d\n", config.RequestCount)
			fmt.Printf("Requests completed, total time: %v\n", duration)
		}

		if mode == "enumerate" {
			seq := probe.NewSequence(runCfg.BaseURL, charset, pathLength)
			startTime := time.Now()
			stats := probe.RunEnumerate(ctx, runCfg, seq, emit)
			duration := time.Since(startTime)
			fmt.Printf("\n统计结果:\n")
			fmt.Printf("成功请求: %d\n", stats.Success)
			fmt.Printf("失败请求: %d\n", stats.Failure)
			fmt.Printf("枚举总数: %d\n", seq.TotalCount())
			fmt.Printf("Requests completed, total time: %v\n", duration)
		}
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
	Short:   "Show application version",
	Aliases: []string{"v"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Version: 1.0.0")
	},
}

var completionCmd = &cobra.Command{
	Use:    "completion [bash|zsh|fish|powershell]",
	Short:  "Generate completion script",
	Hidden: true,
	Long: `To load completions:

Bash:
  $ source <(ShrtProbe completion bash)

Zsh:
  $ ShrtProbe completion zsh > "${fpath[1]}/_ShrtProbe"

Fish:
  $ ShrtProbe completion fish | source

PowerShell:
  PS> ShrtProbe completion powershell | Out-String | Invoke-Expression
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

func init() {
	rootCmd.Flags().StringVarP(&config.BaseURL, "url", "u", "", "Target URL (required)")
	rootCmd.Flags().StringVarP(&charset, "charset", "s", "", "Character set (default: abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789)")
	rootCmd.Flags().StringVarP(&mode, "mode", "m", "random", "Mode: random or enumerate")
	rootCmd.Flags().IntVarP(&pathLength, "length", "l", 5, "Path length (default: 5)")
	rootCmd.Flags().Int64VarP(&config.RequestCount, "count", "c", 10, "Total request count")
	rootCmd.Flags().IntVarP(&config.Concurrency, "concurrency", "n", 15, "Concurrent workers")
	rootCmd.Flags().DurationVarP(&config.Timeout, "timeout", "t", 10*time.Second, "Per-request timeout")
	rootCmd.Flags().StringVarP(&proxyAddr, "proxy", "p", "", "Proxy server address (format: http://host:port)")
	rootCmd.Flags().StringSliceVar(&failBody, "fail-body", nil, "Failure if response body contains substring (case-insensitive, repeatable)")
	rootCmd.Flags().IntSliceVar(&failStatus, "fail-status", nil, "Failure if HTTP status equals value (repeatable)")
	rootCmd.MarkFlagRequired("url")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(completionCmd)
}
