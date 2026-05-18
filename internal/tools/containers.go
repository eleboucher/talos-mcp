package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/siderolabs/talos/pkg/machinery/api/common"
	machineapi "github.com/siderolabs/talos/pkg/machinery/api/machine"
	talosgo "github.com/siderolabs/talos/pkg/machinery/client"

	"git.erwanleboucher.dev/eleboucher/talos-mcp/internal/talosclient"
)

func driverFromArg(s string) common.ContainerDriver {
	if s == "cri" {
		return common.ContainerDriver_CRI
	}
	return common.ContainerDriver_CONTAINERD
}

func registerContainers(s *server.MCPServer, f *talosclient.Factory) {
	opts := append([]mcp.ToolOption{
		mcp.WithDescription("List containers running on the node (equivalent to `talosctl containers`)."),
		mcp.WithString("namespace",
			mcp.Description("Containerd namespace. Defaults to `system`."),
			mcp.DefaultString("system"),
		),
		mcp.WithString("driver",
			mcp.Description("Container driver."),
			mcp.Enum("containerd", "cri"),
		),
	}, commonArgs()...)
	tool := mcp.NewTool("talos_containers", opts...)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		c, callCtx, err := f.New(ctx, parseOverride(req))
		if err != nil {
			return errResult(err), nil
		}
		defer func() { _ = c.Close() }()
		ns := req.GetString("namespace", "system")
		drv := driverFromArg(req.GetString("driver", ""))
		resp, err := runResult(ctx, c, func(tc *talosgo.Client) (*machineapi.ContainersResponse, error) {
			return tc.Containers(callCtx, ns, drv)
		})
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(resp)
	})
}

func registerContainerStats(s *server.MCPServer, f *talosclient.Factory) {
	opts := append([]mcp.ToolOption{
		mcp.WithDescription("Container resource stats (equivalent to `talosctl stats`)."),
		mcp.WithString("namespace",
			mcp.Description("Containerd namespace. Defaults to `system`."),
			mcp.DefaultString("system"),
		),
		mcp.WithString("driver",
			mcp.Description("Container driver."),
			mcp.Enum("containerd", "cri"),
		),
	}, commonArgs()...)
	tool := mcp.NewTool("talos_container_stats", opts...)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		c, callCtx, err := f.New(ctx, parseOverride(req))
		if err != nil {
			return errResult(err), nil
		}
		defer func() { _ = c.Close() }()
		ns := req.GetString("namespace", "system")
		drv := driverFromArg(req.GetString("driver", ""))
		resp, err := runResult(ctx, c, func(tc *talosgo.Client) (*machineapi.StatsResponse, error) {
			return tc.Stats(callCtx, ns, drv)
		})
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(resp)
	})
}
