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
	requestCount   int64
	defaultCharset = "abcdefghijklmnopqrstuvwxyz0123456789"
	charset        string
	mode           string
	pathLength     int
	concurrency    int
	addColor       = color.New(color.FgGreen).Add(color.Bold).PrintfFunc()
	removeColor    = color.New(color.FgRed).Add(color.Bold).PrintfFunc()
	errorColor     = color.New(color.FgRed).Add(color.Bold).PrintfFunc()
	config         = RequestConfig{
		URL:          "",
		Timeout:      10 * time.Second,
		RequestCount: 100,
		Concurrency:  10,
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

		proxy, err := Proxy(proxyAddr)
		if err != nil {
			removeColor("Proxy setup error: %v\n", err)
		} else {
			httpClient = *proxy
			addColor("Proxy set successfully: %s\n", proxyAddr)
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

		if mode == "random" {
			var err error
			config.URL, err = GenerateRandomURL(config.URL, charset, pathLength)
			if err != nil {
				errorColor("Failed to generate random URL: %v\n", err)
				os.Exit(1)
			}
			addColor("Generated random URL: %s\n", config.URL)
			sendRequestsConcurrently(config)
		}

		if mode == "enumerate" {
			generator = NewURLGenerator(config.URL, charset, pathLength)
			sendRequestsConcurrentlyWithGenerator(config, generator)
		}

		startTime := time.Now()
		duration := time.Since(startTime)
		fmt.Printf("Requests completed, total time: %v\n", duration)
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
	rootCmd.Flags().StringVarP(&charset, "charset", "s", "", "Character set (default: abcdefghijklmnopqrstuvwxyz0123456789)")
	rootCmd.Flags().StringVarP(&mode, "mode", "m", "random", "Mode: random or enumerate")
	rootCmd.Flags().IntVarP(&pathLength, "length", "l", 5, "Path length (default: 5)")
	rootCmd.Flags().Int64VarP(&config.RequestCount, "count", "c", 100, "Total request count (default: 100)")
	rootCmd.Flags().IntVarP(&config.Concurrency, "concurrency", "n", 10, "Concurrency level (default: 10)")
	rootCmd.Flags().StringVarP(&proxyAddr, "proxy", "p", "", "Proxy server address (format: http://host:port)")
	rootCmd.MarkFlagRequired("url")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(completionCmd)
}
