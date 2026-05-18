package tools

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	clusterapi "github.com/siderolabs/talos/pkg/machinery/api/cluster"

	"git.erwanleboucher.dev/eleboucher/talos-mcp/internal/talosclient"
)

func registerHealth(s *server.MCPServer, f *talosclient.Factory) {
	opts := append([]mcp.ToolOption{
		mcp.WithDescription("Run a Talos cluster health check (equivalent to `talosctl health`). Returns the streamed progress messages."),
		mcp.WithArray("control_plane_nodes",
			mcp.Description("Override the set of control-plane node IPs/hostnames to check."),
			mcp.WithStringItems(),
		),
		mcp.WithArray("worker_nodes",
			mcp.Description("Override the set of worker node IPs/hostnames to check."),
			mcp.WithStringItems(),
		),
		mcp.WithNumber("wait_timeout_seconds",
			mcp.Description("How long to wait for the cluster to become healthy. Defaults to 60."),
		),
	}, commonArgs()...)
	tool := mcp.NewTool("talos_health", opts...)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		c, callCtx, err := f.New(ctx, parseOverride(req))
		if err != nil {
			return errResult(err), nil
		}
		defer func() { _ = c.Close() }()

		timeout := time.Duration(req.GetInt("wait_timeout_seconds", 60)) * time.Second
		if deadline, ok := callCtx.Deadline(); ok {
			if budget := time.Until(deadline); budget > 0 && timeout > budget {
				slog.Warn("health wait_timeout_seconds exceeds tool budget; clamping",
					"requested", timeout, "budget", budget)
				timeout = budget
			}
		}
		info := &clusterapi.ClusterInfo{
			ControlPlaneNodes: req.GetStringSlice("control_plane_nodes", nil),
			WorkerNodes:       req.GetStringSlice("worker_nodes", nil),
		}
		stream, err := c.Talos().ClusterHealthCheck(callCtx, timeout, info)
		if err != nil {
			return errResult(err), nil
		}
		var sb strings.Builder
		var streamErr error
		for {
			msg, recvErr := stream.Recv()
			if errors.Is(recvErr, io.EOF) {
				break
			}
			if recvErr != nil {
				streamErr = recvErr
				break
			}
			if errStr := msg.GetMetadata().GetError(); errStr != "" {
				sb.WriteString("NODE ERROR: " + errStr + "\n")
			}
			if msg.GetMessage() != "" {
				sb.WriteString(msg.GetMessage())
				sb.WriteByte('\n')
			}
		}
		if streamErr != nil {
			return errResult(fmt.Errorf("health stream aborted: %w\npartial output:\n%s",
				streamErr, sb.String())), nil
		}
		return mcp.NewToolResultText(sb.String()), nil
	})
}
