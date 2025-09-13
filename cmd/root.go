/*
Copyright © 2025 Gi1gamesh666 <208263442@qq.com>
*/
package cmd

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	httpClient     http.Client = http.Client{}
	proxyAddr      string
	defaultCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charset        string
	mode           string
	pathLength     int
	addColor       = color.New(color.FgGreen).Add(color.Bold).PrintfFunc()
	removeColor    = color.New(color.FgRed).Add(color.Bold).PrintfFunc()
	errorColor     = color.New(color.FgRed).Add(color.Bold).PrintfFunc()
	config         = RequestConfig{
		httpClient:   httpClient,
		URL:          "",
		Timeout:      10 * time.Second,
		RequestCount: 100,
		Concurrency:  5,
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
	generator *URLGenerator
)

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
  ShrtProbe proxy --proxy http://proxy.example.com:8080
  ShrtProbe --url http://example.com/`,

	Run: func(cmd *cobra.Command, args []string) {
		// Check URL
		if config.URL == "" {
			errorColor("Error: Target URL is required\n")
			os.Exit(1)
		}

		// Check charset
		if charset == "" {
			addColor("Using default charset: %s\n", defaultCharset)
			charset = defaultCharset
		} else {
			charset = removeDuplicates(charset)
			addColor("Charset set: %s\n", charset)
		}

		if proxyAddr != "" {
			proxy, err := Proxy(proxyAddr)
			if err != nil {
				removeColor("Proxy setup error: %v\n", err)
			} else {
				httpClient = *proxy
				addColor("Proxy set successfully: %s\n", proxyAddr)
			}
		}

		// Check mode
		if mode != "enumerate" && mode != "random" {
			removeColor("Error: Unknown mode. Use 'enumerate' or 'random'\n")
			os.Exit(1)
		}

		// Check path length
		if pathLength <= 0 {
			removeColor("Error: Path length must be greater than 0\n")
			os.Exit(1)
		}

		fmt.Println("")
		addColor("Target URL:     %s\n", config.URL)
		addColor("Mode:           %s\n", mode)
		addColor("Charset:        %s\n", charset)
		addColor("Path Length:    %d\n", pathLength)
		addColor("Request Count:  %d\n", config.RequestCount)
		addColor("Concurrency:    %d\n", config.Concurrency)

		if mode == "random" {
			if config.URL[len(config.URL)-1] != '/' {
				config.URL += "/"
			}
			var err error
			if err != nil {
				errorColor("Failed to generate random URL: %v\n", err)
				os.Exit(1)
			}

			startTime := time.Now()
			sendRequestsConcurrently(config)
			duration := time.Since(startTime)
			fmt.Printf("Requests completed, total time: %v\n", duration)
		}

		if mode == "enumerate" {
			generator = NewURLGenerator(config.URL, charset, pathLength)
			startTime := time.Now()
			sendRequestsConcurrentlyWithGenerator(config, generator)
			duration := time.Since(startTime)
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
	rootCmd.Flags().StringVarP(&config.URL, "url", "u", "", "Target URL (required)")
	rootCmd.Flags().StringVarP(&charset, "charset", "s", "", "Character set (default: abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789)")
	rootCmd.Flags().StringVarP(&mode, "mode", "m", "random", "Mode: random or enumerate")
	rootCmd.Flags().IntVarP(&pathLength, "length", "l", 5, "Path length (default: 5)")
	rootCmd.Flags().Int64VarP(&config.RequestCount, "count", "c", 10, "Total request count (default: 100)")
	rootCmd.Flags().IntVarP(&config.Concurrency, "concurrency", "n", 15, "Concurrency level (default: 10)")
	rootCmd.Flags().StringVarP(&proxyAddr, "proxy", "p", "", "Proxy server address (format: http://host:port)")
	rootCmd.MarkFlagRequired("url")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(completionCmd)
}
