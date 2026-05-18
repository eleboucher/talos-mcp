package tools

import (
	"context"
	"errors"
	"io"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	machineapi "github.com/siderolabs/talos/pkg/machinery/api/machine"
	storageapi "github.com/siderolabs/talos/pkg/machinery/api/storage"
	talosgo "github.com/siderolabs/talos/pkg/machinery/client"
	"google.golang.org/protobuf/types/known/emptypb"

	"git.erwanleboucher.dev/eleboucher/talos-mcp/internal/talosclient"
)

func registerDisks(s *server.MCPServer, f *talosclient.Factory) {
	opts := append([]mcp.ToolOption{
		mcp.WithDescription("List block devices on one or more nodes (equivalent to `talosctl disks`)."),
	}, commonArgs()...)
	tool := mcp.NewTool("talos_disks", opts...)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		c, callCtx, err := f.New(ctx, parseOverride(req))
		if err != nil {
			return errResult(err), nil
		}
		defer func() { _ = c.Close() }()

		resp, err := runResult(ctx, c, func(tc *talosgo.Client) (*storageapi.DisksResponse, error) {
			return tc.Disks(callCtx)
		})
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(resp)
	})
}

func registerDiskStats(s *server.MCPServer, f *talosclient.Factory) {
	opts := append([]mcp.ToolOption{
		mcp.WithDescription("Per-device block I/O counters from /proc/diskstats (equivalent to `talosctl diskstats`)."),
	}, commonArgs()...)
	tool := mcp.NewTool("talos_disk_stats", opts...)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		c, callCtx, err := f.New(ctx, parseOverride(req))
		if err != nil {
			return errResult(err), nil
		}
		defer func() { _ = c.Close() }()
		resp, err := runResult(ctx, c, func(tc *talosgo.Client) (*machineapi.DiskStatsResponse, error) {
			return tc.MachineClient.DiskStats(callCtx, &emptypb.Empty{})
		})
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(resp)
	})
}

func registerDiskUsage(s *server.MCPServer, f *talosclient.Factory) {
	opts := append([]mcp.ToolOption{
		mcp.WithDescription("Report disk usage for one or more paths (equivalent to `talosctl usage`)."),
		mcp.WithArray("paths",
			mcp.Required(),
			mcp.Description("Paths to measure (e.g. /var/lib)."),
			mcp.WithStringItems(),
		),
		mcp.WithNumber("recursion_depth",
			mcp.Description("How many levels of subdirectories to recurse. 0 = unlimited (default)."),
		),
		mcp.WithBoolean("all",
			mcp.Description("Include regular files, not just directories."),
		),
	}, commonArgs()...)
	tool := mcp.NewTool("talos_disk_usage", opts...)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		paths, err := req.RequireStringSlice("paths")
		if err != nil {
			return errResult(err), nil
		}
		c, callCtx, err := f.New(ctx, parseOverride(req))
		if err != nil {
			return errResult(err), nil
		}
		defer func() { _ = c.Close() }()

		stream, err := c.Talos().DiskUsage(callCtx, &machineapi.DiskUsageRequest{
			Paths:          paths,
			RecursionDepth: toInt32(req.GetInt("recursion_depth", 0)),
			All:            req.GetBool("all", false),
		})
		if err != nil {
			return errResult(err), nil
		}
		var infos []*machineapi.DiskUsageInfo
		for {
			info, recvErr := stream.Recv()
			if errors.Is(recvErr, io.EOF) {
				break
			}
			if recvErr != nil {
				return errResult(recvErr), nil
			}
			infos = append(infos, info)
		}
		return jsonArrayResult(infos)
	})
}
