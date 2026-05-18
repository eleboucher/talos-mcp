package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	machineapi "github.com/siderolabs/talos/pkg/machinery/api/machine"
	talosgo "github.com/siderolabs/talos/pkg/machinery/client"

	"git.erwanleboucher.dev/eleboucher/talos-mcp/internal/talosclient"
)

func registerNetstat(s *server.MCPServer, f *talosclient.Factory) {
	opts := append([]mcp.ToolOption{
		mcp.WithDescription("Show network connections (equivalent to `talosctl netstat`)."),
		mcp.WithString("l4proto",
			mcp.Description("Layer-4 protocol filter."),
			mcp.Enum("TCP", "UDP", "RAW", "TCP6", "UDP6", "RAW6"),
		),
		mcp.WithString("filter",
			mcp.Description("Connection state filter."),
			mcp.Enum("ALL", "CONNECTED", "LISTENING"),
		),
	}, commonArgs()...)
	tool := mcp.NewTool("talos_netstat", opts...)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		c, callCtx, err := f.New(ctx, parseOverride(req))
		if err != nil {
			return errResult(err), nil
		}
		defer func() { _ = c.Close() }()

		nrq := &machineapi.NetstatRequest{
			Feature: &machineapi.NetstatRequest_Feature{Pid: true},
		}
		switch req.GetString("l4proto", "") {
		case "TCP":
			nrq.L4Proto = &machineapi.NetstatRequest_L4Proto{Tcp: true}
		case "TCP6":
			nrq.L4Proto = &machineapi.NetstatRequest_L4Proto{Tcp6: true}
		case "UDP":
			nrq.L4Proto = &machineapi.NetstatRequest_L4Proto{Udp: true}
		case "UDP6":
			nrq.L4Proto = &machineapi.NetstatRequest_L4Proto{Udp6: true}
		case "RAW":
			nrq.L4Proto = &machineapi.NetstatRequest_L4Proto{Raw: true}
		case "RAW6":
			nrq.L4Proto = &machineapi.NetstatRequest_L4Proto{Raw6: true}
		}
		switch req.GetString("filter", "") {
		case "CONNECTED":
			nrq.Filter = machineapi.NetstatRequest_CONNECTED
		case "LISTENING":
			nrq.Filter = machineapi.NetstatRequest_LISTENING
		default:
			nrq.Filter = machineapi.NetstatRequest_ALL
		}
		resp, err := runResult(ctx, c, func(tc *talosgo.Client) (*machineapi.NetstatResponse, error) {
			return tc.Netstat(callCtx, nrq)
		})
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(resp)
	})
}
