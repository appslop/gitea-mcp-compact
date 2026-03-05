package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"gitea.com/gitea/gitea-mcp/operation"
	flagPkg "gitea.com/gitea/gitea-mcp/pkg/flag"
	"gitea.com/gitea/gitea-mcp/pkg/log"
	"gitea.com/gitea/gitea-mcp/pkg/settings"
)

var (
	host        string
	port        int
	token       string
	userToken string
	version     bool
)

func init() {
	flag.StringVar(&flagPkg.Mode, "t", "stdio", "")
	flag.StringVar(&flagPkg.Mode, "transport", "stdio", "")
	flag.StringVar(&host, "H", os.Getenv("GITEA_HOST"), "")
	flag.StringVar(&host, "host", os.Getenv("GITEA_HOST"), "")
	flag.IntVar(&port, "p", 8080, "")
	flag.IntVar(&port, "port", 8080, "")
	flag.StringVar(&token, "T", "", "")
	flag.StringVar(&token, "token", "", "")
	flag.StringVar(&userToken, "ut", "", "")
	flag.StringVar(&userToken, "user-token", "", "")
	flag.BoolVar(&flagPkg.ReadOnly, "r", false, "")
	flag.BoolVar(&flagPkg.ReadOnly, "read-only", false, "")
	flag.BoolVar(&flagPkg.Debug, "d", false, "")
	flag.BoolVar(&flagPkg.Debug, "debug", false, "")
	flag.BoolVar(&flagPkg.Insecure, "k", false, "")
	flag.BoolVar(&flagPkg.Insecure, "insecure", false, "")
	flag.BoolVar(&version, "v", false, "")
	flag.BoolVar(&version, "version", false, "")

	flag.Usage = func() {
		w := tabwriter.NewWriter(os.Stderr, 0, 0, 3, ' ', 0)
		fmt.Fprintln(os.Stderr, "Usage: gitea-mcp [options]")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Options:")
		fmt.Fprintf(w, "  -t, -transport <type>\tTransport type: stdio or http (default: stdio)\n")
		fmt.Fprintf(w, "  -H, -host <url>\tGitea host URL (default: https://gitea.com)\n")
		fmt.Fprintf(w, "  -p, -port <number>\tHTTP server port (default: 8080)\n")
		fmt.Fprintf(w, "  -T, -token <token>\tPersonal access token\n")
		fmt.Fprintf(w, "  -ut, -user-token <token>\tUser token with elevated permissions\n")
		fmt.Fprintf(w, "  -r, -read-only\tExpose only read-only tools\n")
		fmt.Fprintf(w, "  -d, -debug\tEnable debug mode\n")
		fmt.Fprintf(w, "  -k, -insecure\tIgnore TLS certificate errors\n")
		fmt.Fprintf(w, "  -v, -version\tPrint version and exit\n")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Environment variables:")
		fmt.Fprintf(w, "  GITEA_ACCESS_TOKEN\tProvide access token\n")
		fmt.Fprintf(w, "  GITEA_USER_ACCESS_TOKEN\tProvide user token with elevated permissions\n")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Settings file:")
		fmt.Fprintf(w, "  ~/.gitea-mcp/settings.json\tPersistent configuration (JSON)\n")
		fmt.Fprintf(w, "    {\n")
		fmt.Fprintf(w, "      \"token\": \"your-user-token-here\",\n")
		fmt.Fprintf(w, "      \"user_token\": \"your-user-token-here\",\n")
		fmt.Fprintf(w, "      \"host\": \"https://gitea.com\"\n")
		fmt.Fprintf(w, "    }\n")
		fmt.Fprintln(w)
		fmt.Fprintf(w, "  GITEA_DEBUG\tSet to 'true' for debug mode\n")
		fmt.Fprintf(w, "  GITEA_HOST\tOverride Gitea host URL\n")
		fmt.Fprintf(w, "  GITEA_INSECURE\tSet to 'true' to ignore TLS errors\n")
		fmt.Fprintf(w, "  GITEA_READONLY\tSet to 'true' for read-only mode\n")
		fmt.Fprintf(w, "  MCP_MODE\tOverride transport mode\n")
		w.Flush()
	}

	flag.Parse()

	flagPkg.Host = host
	if flagPkg.Host == "" {
		flagPkg.Host = "https://gitea.com"
	}

	flagPkg.Port = port

	// Load settings file
	settingsConfig := settings.Load()
	log.Debugf("Settings loaded: agentToken=%t, userToken=%t, host=%s",
		settingsConfig.Token != "", settingsConfig.UserToken != "", settingsConfig.Host)

	// Host precedence: CLI flag > settings file > default
	if host == "" && settingsConfig.Host != "" {
		flagPkg.Host = settingsConfig.Host
		log.Debugf("Using host from settings file: %s", settingsConfig.Host)
	}

	// Agent token precedence: CLI flag > settings file > env var
	flagPkg.Token = token
	if flagPkg.Token == "" {
		if settingsConfig.Token != "" {
			flagPkg.Token = settingsConfig.Token
			log.Debugf("Using agent token from settings file")
		}
	}
	if flagPkg.Token == "" {
		flagPkg.Token = os.Getenv("GITEA_ACCESS_TOKEN")
		if flagPkg.Token != "" {
			log.Debugf("Using agent token from environment variable")
		}
	}
	log.Debugf("Final agent token: empty=%t", flagPkg.Token == "")

	// System token precedence: CLI flag > settings file > env var
	flagPkg.UserToken = userToken
	if flagPkg.UserToken == "" {
		if settingsConfig.UserToken != "" {
			flagPkg.UserToken = settingsConfig.UserToken
		}
	}
	if flagPkg.UserToken == "" {
		flagPkg.UserToken = os.Getenv("GITEA_USER_ACCESS_TOKEN")
	}

	// CreateRepoAsUser setting: settings file > default (false)
	flagPkg.CreateRepoAsUser = settingsConfig.CreateRepoAsUser
	log.Debugf("CreateRepoAsUser: %t", flagPkg.CreateRepoAsUser)

	// TruncateCompact setting: settings file > default (100)
	flagPkg.TruncateCompact = settingsConfig.TruncateCompact
	if flagPkg.TruncateCompact == 0 {
		flagPkg.TruncateCompact = 100
	}
	log.Debugf("TruncateCompact: %d", flagPkg.TruncateCompact)

	// TruncateFull setting: settings file > default (300)
	flagPkg.TruncateFull = settingsConfig.TruncateFull
	if flagPkg.TruncateFull == 0 {
		flagPkg.TruncateFull = 300
	}
	log.Debugf("TruncateFull: %d", flagPkg.TruncateFull)

	if os.Getenv("MCP_MODE") != "" {
		flagPkg.Mode = os.Getenv("MCP_MODE")
	}

	if os.Getenv("GITEA_READONLY") == "true" {
		flagPkg.ReadOnly = true
	}

	if os.Getenv("GITEA_DEBUG") == "true" {
		flagPkg.Debug = true
	}

	// Set insecure mode based on environment variable
	if os.Getenv("GITEA_INSECURE") == "true" {
		flagPkg.Insecure = true
	}
}

func Execute() {
	if version {
		fmt.Fprintln(os.Stdout, flagPkg.Version)
		return
	}
	defer log.Default().Sync() //nolint:errcheck // best-effort flush
	if err := operation.Run(); err != nil {
		if err == context.Canceled {
			log.Info("Server shutdown due to context cancellation")
			return
		}
		log.Fatalf("Run Gitea MCP Server Error: %v", err) //nolint:gocritic // intentional exit after defer
	}
}
