package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	machineapi "github.com/siderolabs/talos/pkg/machinery/api/machine"
	talosgo "github.com/siderolabs/talos/pkg/machinery/client"

	"git.erwanleboucher.dev/eleboucher/talos-mcp/internal/talosclient"
)

func registerEtcdMembers(s *server.MCPServer, f *talosclient.Factory) {
	opts := append([]mcp.ToolOption{
		mcp.WithDescription("List etcd cluster members (equivalent to `talosctl etcd members`)."),
		mcp.WithBoolean("query_local",
			mcp.Description("Query the local node's view instead of the cluster's."),
		),
	}, commonArgs()...)
	tool := mcp.NewTool("talos_etcd_members", opts...)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		c, callCtx, err := f.New(ctx, parseOverride(req))
		if err != nil {
			return errResult(err), nil
		}
		defer func() { _ = c.Close() }()
		ql := req.GetBool("query_local", false)
		resp, err := runResult(ctx, c, func(tc *talosgo.Client) (*machineapi.EtcdMemberListResponse, error) {
			return tc.EtcdMemberList(callCtx, &machineapi.EtcdMemberListRequest{QueryLocal: ql})
		})
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(resp)
	})
}

func registerEtcdStatus(s *server.MCPServer, f *talosclient.Factory) {
	opts := append([]mcp.ToolOption{
		mcp.WithDescription("Show etcd status (leader, raft index, db size — equivalent to `talosctl etcd status`)."),
	}, commonArgs()...)
	tool := mcp.NewTool("talos_etcd_status", opts...)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		c, callCtx, err := f.New(ctx, parseOverride(req))
		if err != nil {
			return errResult(err), nil
		}
		defer func() { _ = c.Close() }()
		resp, err := runResult(ctx, c, func(tc *talosgo.Client) (*machineapi.EtcdStatusResponse, error) {
			return tc.EtcdStatus(callCtx)
		})
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(resp)
	})
}

func registerEtcdAlarms(s *server.MCPServer, f *talosclient.Factory) {
	opts := append([]mcp.ToolOption{
		mcp.WithDescription("List active etcd alarms (equivalent to `talosctl etcd alarm list`). An empty list means etcd is healthy."),
	}, commonArgs()...)
	tool := mcp.NewTool("talos_etcd_alarms", opts...)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		c, callCtx, err := f.New(ctx, parseOverride(req))
		if err != nil {
			return errResult(err), nil
		}
		defer func() { _ = c.Close() }()
		resp, err := runResult(ctx, c, func(tc *talosgo.Client) (*machineapi.EtcdAlarmListResponse, error) {
			return tc.EtcdAlarmList(callCtx)
		})
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(resp)
	})
}
