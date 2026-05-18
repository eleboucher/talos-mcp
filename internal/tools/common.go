package tools

import (
	"context"
	"fmt"
	"log/slog"
	"math"

	"github.com/mark3labs/mcp-go/mcp"
	talosgo "github.com/siderolabs/talos/pkg/machinery/client"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"

	"git.erwanleboucher.dev/eleboucher/talos-mcp/internal/talosclient"
)

func toInt32(i int) int32 {
	switch {
	case i < 0:
		return 0
	case i > math.MaxInt32:
		return math.MaxInt32
	default:
		return int32(i)
	}
}

func runResult[T any](ctx context.Context, c *talosclient.Client, op func(tc *talosgo.Client) (T, error)) (T, error) {
	var result T
	err := c.Run(ctx, func(tc *talosgo.Client) error {
		var err error
		result, err = op(tc)
		return err
	})
	return result, err
}

func commonArgs() []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithArray("nodes",
			mcp.Description("Target node IPs/hostnames. If omitted, the default node(s) from talosconfig are used."),
			mcp.WithStringItems(),
		),
		mcp.WithString("context",
			mcp.Description("talosconfig context name."),
		),
		mcp.WithArray("endpoints",
			mcp.Description("Override talosconfig endpoints."),
			mcp.WithStringItems(),
		),
	}
}

func parseOverride(req mcp.CallToolRequest) talosclient.Override {
	return talosclient.Override{
		Context:   req.GetString("context", ""),
		Endpoints: req.GetStringSlice("endpoints", nil),
		Nodes:     req.GetStringSlice("nodes", nil),
	}
}

func errResult(err error) *mcp.CallToolResult {
	return mcp.NewToolResultError(err.Error())
}

//nolint:unparam // error slot kept so handlers can `return jsonResult(resp)`
func jsonResult(m protoreflect.ProtoMessage) (*mcp.CallToolResult, error) {
	b, err := protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(m)
	if err != nil {
		slog.Error("marshal response", "type", fmt.Sprintf("%T", m), "err", err)
		return errResult(fmt.Errorf("marshal response: %w", err)), nil
	}
	return mcp.NewToolResultText(string(b)), nil
}

func jsonArrayResult[T protoreflect.ProtoMessage](items []T) (*mcp.CallToolResult, error) {
	opts := protojson.MarshalOptions{Multiline: false, Indent: ""}
	var b []byte
	b = append(b, '[')
	for i, it := range items {
		if i > 0 {
			b = append(b, ',', '\n')
		} else {
			b = append(b, '\n')
		}
		chunk, err := opts.Marshal(it)
		if err != nil {
			slog.Error("marshal item", "index", i, "type", fmt.Sprintf("%T", it), "err", err)
			return errResult(fmt.Errorf("marshal item %d: %w", i, err)), nil
		}
		b = append(b, ' ', ' ')
		b = append(b, chunk...)
	}
	b = append(b, '\n', ']')
	return mcp.NewToolResultText(string(b)), nil
}
