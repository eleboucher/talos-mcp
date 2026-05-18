package main

import (
	"context"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"git.erwanleboucher.dev/eleboucher/talos-mcp/internal/talosclient"
	"git.erwanleboucher.dev/eleboucher/talos-mcp/internal/tools"
)

var (
	version = "dev"
	commit  = "unknown"
)

const defaultToolTimeout = 20 * time.Second

func main() {
	setupLogger()

	f, err := talosclient.NewFactory()
	if err != nil {
		slog.Error("init talos client factory", "err", err)
		os.Exit(1)
	}

	timeout := toolTimeout()
	slog.Info("talos-mcp starting",
		"version", version, "commit", commit, "tool_timeout", timeout)
	s := server.NewMCPServer("talos-mcp", version,
		server.WithToolCapabilities(true),
		server.WithRecovery(),
		server.WithToolHandlerMiddleware(loggingMiddleware(timeout)),
	)
	tools.RegisterAll(s, f)

	transport := os.Getenv("TALOS_MCP_TRANSPORT")
	if transport == "" {
		transport = "stdio"
	}
	slog.Info("serving", "transport", transport)
	switch transport {
	case "stdio":
		if err := server.ServeStdio(s); err != nil {
			slog.Error("stdio serve", "err", err)
			os.Exit(1)
		}
	case "http":
		addr := os.Getenv("TALOS_MCP_HTTP_ADDR")
		if addr == "" {
			addr = ":8080"
		}
		slog.Info("http listening", "addr", addr)
		hs := server.NewStreamableHTTPServer(s)
		if err := hs.Start(addr); err != nil {
			slog.Error("http serve", "err", err)
			os.Exit(1)
		}
	default:
		slog.Error("unknown transport", "value", transport, "want", "stdio|http")
		os.Exit(1)
	}
}

func setupLogger() {
	level := slog.LevelInfo
	switch strings.ToLower(os.Getenv("TALOS_MCP_LOG_LEVEL")) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(h))
}

func toolTimeout() time.Duration {
	v := os.Getenv("TALOS_MCP_TOOL_TIMEOUT")
	if v == "" {
		return defaultToolTimeout
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		slog.Warn("invalid TALOS_MCP_TOOL_TIMEOUT, using default",
			"value", v, "default", defaultToolTimeout, "err", err)
		return defaultToolTimeout
	}
	return d
}

func loggingMiddleware(timeout time.Duration) server.ToolHandlerMiddleware {
	return func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, req mcp.CallToolRequest) (res *mcp.CallToolResult, err error) {
			name := req.Params.Name
			start := time.Now()
			slog.Info("tool start", "tool", name, "timeout", timeout)
			slog.Debug("tool args", "tool", name, "args", req.GetArguments())

			defer func() {
				if r := recover(); r != nil {
					slog.Error("tool panic",
						"tool", name,
						"duration", time.Since(start),
						"panic", r,
						"stack", string(debug.Stack()))
					panic(r)
				}
			}()

			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			res, err = next(ctx, req)
			dur := time.Since(start)
			switch {
			case err != nil:
				slog.Error("tool end", "tool", name, "duration", dur, "err", err)
			case res != nil && res.IsError:
				slog.Warn("tool end", "tool", name, "duration", dur, "result_error", true)
			default:
				slog.Info("tool end", "tool", name, "duration", dur)
			}
			return res, err
		}
	}
}
