package tools

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/siderolabs/talos/pkg/machinery/api/common"

	"git.erwanleboucher.dev/eleboucher/talos-mcp/internal/talosclient"
)

const maxLogLines = 1000

func registerServiceLogs(s *server.MCPServer, f *talosclient.Factory) {
	opts := append([]mcp.ToolOption{
		mcp.WithDescription("Tail logs for a Talos service or container (equivalent to `talosctl logs`)."),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Service or container id (e.g. `kubelet`, `etcd`, `apid`)."),
		),
		mcp.WithString("namespace",
			mcp.Description("Containerd namespace. Defaults to `system`."),
			mcp.DefaultString("system"),
		),
		mcp.WithString("driver",
			mcp.Description("Container driver. Defaults to `containerd`."),
			mcp.Enum("containerd", "cri"),
		),
		mcp.WithNumber("tail_lines",
			mcp.Description("Tail this many lines. 0 = all."),
		),
	}, commonArgs()...)
	tool := mcp.NewTool("talos_service_logs", opts...)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return errResult(err), nil
		}
		ns := req.GetString("namespace", "system")
		driver := common.ContainerDriver_CONTAINERD
		if req.GetString("driver", "") == "cri" {
			driver = common.ContainerDriver_CRI
		}
		c, callCtx, err := f.New(ctx, parseOverride(req))
		if err != nil {
			return errResult(err), nil
		}
		defer func() { _ = c.Close() }()

		callCtx, cancel := context.WithCancel(callCtx)
		defer cancel()

		stream, err := c.Talos().Logs(callCtx, ns, driver, id, false, toInt32(req.GetInt("tail_lines", 0)))
		if err != nil {
			return errResult(err), nil
		}
		var sb strings.Builder
		var n int
		for n < maxLogLines {
			msg, recvErr := stream.Recv()
			if errors.Is(recvErr, io.EOF) {
				break
			}
			if recvErr != nil {
				return errResult(recvErr), nil
			}
			if len(msg.GetBytes()) == 0 {
				continue
			}
			sb.Write(msg.GetBytes())
			if msg.GetBytes()[len(msg.GetBytes())-1] != '\n' {
				sb.WriteByte('\n')
			}
			n++
		}
		return mcp.NewToolResultText(sb.String()), nil
	})
}
