package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	talosgo "github.com/siderolabs/talos/pkg/machinery/client"

	"git.erwanleboucher.dev/eleboucher/talos-mcp/internal/talosclient"
)

func registerKubeconfig(s *server.MCPServer, f *talosclient.Factory) {
	opts := append([]mcp.ToolOption{
		mcp.WithDescription("Download the admin kubeconfig from a control-plane node (equivalent to `talosctl kubeconfig -`)."),
	}, commonArgs()...)
	tool := mcp.NewTool("talos_kubeconfig", opts...)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		c, callCtx, err := f.New(ctx, parseOverride(req))
		if err != nil {
			return errResult(err), nil
		}
		defer func() { _ = c.Close() }()

		data, err := runResult(ctx, c, func(tc *talosgo.Client) ([]byte, error) {
			return tc.Kubeconfig(callCtx)
		})
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	})
}
