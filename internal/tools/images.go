package tools

import (
	"context"
	"errors"
	"io"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/siderolabs/talos/pkg/machinery/api/common"
	machineapi "github.com/siderolabs/talos/pkg/machinery/api/machine"

	"git.erwanleboucher.dev/eleboucher/talos-mcp/internal/talosclient"
)

func imageNamespaceFromArg(s string) common.ContainerdNamespace {
	if s == "system" {
		return common.ContainerdNamespace_NS_SYSTEM
	}
	return common.ContainerdNamespace_NS_CRI
}

func registerImageList(s *server.MCPServer, f *talosclient.Factory) {
	opts := append([]mcp.ToolOption{
		mcp.WithDescription("List cached container images on a node (equivalent to `talosctl image list`)."),
		mcp.WithString("namespace",
			mcp.Description("Containerd namespace. Defaults to `cri` (workloads). Use `system` for Talos system images."),
			mcp.Enum("cri", "system"),
			mcp.DefaultString("cri"),
		),
	}, commonArgs()...)
	tool := mcp.NewTool("talos_image_list", opts...)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		c, callCtx, err := f.New(ctx, parseOverride(req))
		if err != nil {
			return errResult(err), nil
		}
		defer func() { _ = c.Close() }()

		callCtx, cancel := context.WithCancel(callCtx)
		defer cancel()

		stream, err := c.Talos().ImageClient.List(callCtx, &machineapi.ImageServiceListRequest{
			Containerd: &common.ContainerdInstance{
				Driver:    common.ContainerDriver_CONTAINERD,
				Namespace: imageNamespaceFromArg(req.GetString("namespace", "cri")),
			},
		})
		if err != nil {
			return errResult(err), nil
		}
		var entries []*machineapi.ImageServiceListResponse
		for {
			entry, recvErr := stream.Recv()
			if errors.Is(recvErr, io.EOF) {
				break
			}
			if recvErr != nil {
				return errResult(recvErr), nil
			}
			entries = append(entries, entry)
		}
		return jsonArrayResult(entries)
	})
}
