package main

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"

	"git.erwanleboucher.dev/eleboucher/talos-mcp/internal/talosclient"
	"git.erwanleboucher.dev/eleboucher/talos-mcp/internal/tools"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	f, err := talosclient.NewFactory()
	if err != nil {
		fmt.Fprintln(os.Stderr, "talos-mcp:", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "talos-mcp version=%s commit=%s\n", version, commit)
	s := server.NewMCPServer("talos-mcp", version,
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)
	tools.RegisterAll(s, f)

	transport := os.Getenv("TALOS_MCP_TRANSPORT")
	if transport == "" {
		transport = "stdio"
	}
	switch transport {
	case "stdio":
		if err := server.ServeStdio(s); err != nil {
			fmt.Fprintln(os.Stderr, "talos-mcp stdio serve:", err)
			os.Exit(1)
		}
	case "http":
		addr := os.Getenv("TALOS_MCP_HTTP_ADDR")
		if addr == "" {
			addr = ":8080"
		}
		hs := server.NewStreamableHTTPServer(s)
		if err := hs.Start(addr); err != nil {
			fmt.Fprintln(os.Stderr, "talos-mcp http serve:", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "talos-mcp: unknown TALOS_MCP_TRANSPORT=%q (want stdio|http)\n", transport)
		os.Exit(1)
	}
}
