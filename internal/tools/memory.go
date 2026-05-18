package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	machineapi "github.com/siderolabs/talos/pkg/machinery/api/machine"
	talosgo "github.com/siderolabs/talos/pkg/machinery/client"

	"git.erwanleboucher.dev/eleboucher/talos-mcp/internal/talosclient"
)

func registerMemory(s *server.MCPServer, f *talosclient.Factory) {
	opts := append([]mcp.ToolOption{
		mcp.WithDescription("Return /proc/meminfo for one or more nodes (equivalent to `talosctl memory`)."),
	}, commonArgs()...)
	tool := mcp.NewTool("talos_memory", opts...)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		c, callCtx, err := f.New(ctx, parseOverride(req))
		if err != nil {
			return errResult(err), nil
		}
		defer func() { _ = c.Close() }()

		resp, err := runResult(ctx, c, func(tc *talosgo.Client) (*machineapi.MemoryResponse, error) {
			return tc.Memory(callCtx)
		})
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(resp)
	})
}
