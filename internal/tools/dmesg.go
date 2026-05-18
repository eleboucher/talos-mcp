package tools

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"git.erwanleboucher.dev/eleboucher/talos-mcp/internal/talosclient"
)

const maxDmesgLines = 1000

func registerDmesg(s *server.MCPServer, f *talosclient.Factory) {
	opts := append([]mcp.ToolOption{
		mcp.WithDescription("Stream kernel ring buffer; collects up to 1000 lines (equivalent to `talosctl dmesg`)."),
		mcp.WithBoolean("tail",
			mcp.Description("Start from the tail of the buffer (recent messages first)."),
		),
	}, commonArgs()...)
	tool := mcp.NewTool("talos_dmesg", opts...)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		c, callCtx, err := f.New(ctx, parseOverride(req))
		if err != nil {
			return errResult(err), nil
		}
		defer func() { _ = c.Close() }()

		callCtx, cancel := context.WithCancel(callCtx)
		defer cancel()

		stream, err := c.Talos().Dmesg(callCtx, false, req.GetBool("tail", false))
		if err != nil {
			return errResult(err), nil
		}
		var sb strings.Builder
		var n int
		for n < maxDmesgLines {
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
