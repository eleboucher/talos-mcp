package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	timeapi "github.com/siderolabs/talos/pkg/machinery/api/time"
	talosgo "github.com/siderolabs/talos/pkg/machinery/client"

	"git.erwanleboucher.dev/eleboucher/talos-mcp/internal/talosclient"
)

func registerTime(s *server.MCPServer, f *talosclient.Factory) {
	opts := append([]mcp.ToolOption{
		mcp.WithDescription("Return the current Talos server time (equivalent to `talosctl time`)."),
	}, commonArgs()...)
	tool := mcp.NewTool("talos_time", opts...)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		c, callCtx, err := f.New(ctx, parseOverride(req))
		if err != nil {
			return errResult(err), nil
		}
		defer func() { _ = c.Close() }()

		resp, err := runResult(ctx, c, func(tc *talosgo.Client) (*timeapi.TimeResponse, error) {
			return tc.Time(callCtx)
		})
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(resp)
	})
}
