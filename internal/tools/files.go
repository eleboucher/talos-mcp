package tools

import (
	"context"
	"errors"
	"io"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	machineapi "github.com/siderolabs/talos/pkg/machinery/api/machine"

	"git.erwanleboucher.dev/eleboucher/talos-mcp/internal/talosclient"
)

const maxListEntries = 2000

func registerList(s *server.MCPServer, f *talosclient.Factory) {
	opts := append([]mcp.ToolOption{
		mcp.WithDescription("List the contents of a directory on a node (equivalent to `talosctl list`). Capped at 2000 entries."),
		mcp.WithString("root",
			mcp.Required(),
			mcp.Description("Directory to list (e.g. `/etc`)."),
		),
		mcp.WithBoolean("recurse",
			mcp.Description("Recurse into subdirectories."),
		),
		mcp.WithNumber("recursion_depth",
			mcp.Description("Max recursion depth. 0 = unlimited."),
		),
	}, commonArgs()...)
	tool := mcp.NewTool("talos_list", opts...)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		root, err := req.RequireString("root")
		if err != nil {
			return errResult(err), nil
		}
		c, callCtx, err := f.New(ctx, parseOverride(req))
		if err != nil {
			return errResult(err), nil
		}
		defer func() { _ = c.Close() }()

		callCtx, cancel := context.WithCancel(callCtx)
		defer cancel()

		stream, err := c.Talos().LS(callCtx, &machineapi.ListRequest{
			Root:           root,
			Recurse:        req.GetBool("recurse", false),
			RecursionDepth: toInt32(req.GetInt("recursion_depth", 0)),
		})
		if err != nil {
			return errResult(err), nil
		}
		var infos []*machineapi.FileInfo
		for len(infos) < maxListEntries {
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

const maxReadBytes = 1 << 20 // 1 MiB cap

func registerRead(s *server.MCPServer, f *talosclient.Factory) {
	opts := append([]mcp.ToolOption{
		mcp.WithDescription("Read a file from a node (equivalent to `talosctl read`). Capped at 1 MiB."),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Absolute file path on the node (e.g. `/etc/resolv.conf`)."),
		),
	}, commonArgs()...)
	tool := mcp.NewTool("talos_read", opts...)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := req.RequireString("path")
		if err != nil {
			return errResult(err), nil
		}
		c, callCtx, err := f.New(ctx, parseOverride(req))
		if err != nil {
			return errResult(err), nil
		}
		defer func() { _ = c.Close() }()

		callCtx, cancel := context.WithCancel(callCtx)
		defer cancel()

		r, err := c.Talos().Read(callCtx, path)
		if err != nil {
			return errResult(err), nil
		}
		defer func() { _ = r.Close() }()

		data, err := io.ReadAll(io.LimitReader(r, maxReadBytes))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	})
}
