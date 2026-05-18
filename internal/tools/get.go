package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cosi-project/runtime/pkg/resource"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	talosgo "github.com/siderolabs/talos/pkg/machinery/client"
	"gopkg.in/yaml.v3"

	"git.erwanleboucher.dev/eleboucher/talos-mcp/internal/talosclient"
)

func registerGet(s *server.MCPServer, f *talosclient.Factory) {
	opts := append([]mcp.ToolOption{
		mcp.WithDescription("Read COSI resources from a Talos node (equivalent to `talosctl get`). Single tool covers all resource types — try `members`, `nodeips`, `routes`, `addresses`, `machineconfigs`, `etcdmembers`, `mountstatuses`, `extensions`, etc. Pass a single node in `nodes`; COSI doesn't support multi-node proxying."),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("Resource type or alias (e.g. `members`, `mc` for machineconfig, `routes`)."),
		),
		mcp.WithString("id",
			mcp.Description("Optional resource ID. If omitted, returns the full list for the type."),
		),
		mcp.WithString("namespace",
			mcp.Description("Resource namespace. Defaults to the type's default namespace."),
		),
		mcp.WithString("output",
			mcp.Description("Output format. `yaml` (default) preserves full metadata + spec, `json` is jq-friendly, `table` is a compact summary."),
			mcp.Enum("yaml", "json", "table"),
			mcp.DefaultString("yaml"),
		),
	}, commonArgs()...)
	tool := mcp.NewTool("talos_get", opts...)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		typ, err := req.RequireString("type")
		if err != nil {
			return errResult(err), nil
		}
		id := req.GetString("id", "")
		ns := req.GetString("namespace", "")
		output := req.GetString("output", "yaml")

		c, callCtx, err := f.NewSingleNode(ctx, parseOverride(req))
		if err != nil {
			return errResult(err), nil
		}
		defer func() { _ = c.Close() }()

		rd, err := c.Talos().ResolveResourceKind(callCtx, &ns, typ)
		if err != nil {
			return errResult(err), nil
		}
		resolvedType := rd.TypedSpec().Type

		var items []resource.Resource
		if id != "" {
			r, err := runResult(ctx, c, func(tc *talosgo.Client) (resource.Resource, error) {
				return tc.COSI.Get(callCtx, resource.NewMetadata(ns, resolvedType, id, resource.VersionUndefined))
			})
			if err != nil {
				return errResult(err), nil
			}
			items = []resource.Resource{r}
		} else {
			items, err = runResult(ctx, c, func(tc *talosgo.Client) ([]resource.Resource, error) {
				list, err := tc.COSI.List(callCtx, resource.NewMetadata(ns, resolvedType, "", resource.VersionUndefined))
				if err != nil {
					return nil, err
				}
				return list.Items, nil
			})
			if err != nil {
				return errResult(err), nil
			}
		}

		switch output {
		case "table":
			return mcp.NewToolResultText(formatResourceTable(items)), nil
		case "json":
			return mcp.NewToolResultText(formatResourcesJSON(items)), nil
		default:
			return mcp.NewToolResultText(formatResourcesYAML(items)), nil
		}
	})
}

func formatResourceTable(items []resource.Resource) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%-20s %-20s %-40s %s\n", "NAMESPACE", "TYPE", "ID", "VERSION")
	for _, r := range items {
		m := r.Metadata()
		fmt.Fprintf(&sb, "%-20s %-20s %-40s %s\n", m.Namespace(), m.Type(), m.ID(), m.Version())
	}
	return sb.String()
}

type resourceEnvelope struct {
	Namespace string `yaml:"namespace" json:"namespace"`
	Type      string `yaml:"type" json:"type"`
	ID        string `yaml:"id" json:"id"`
	Version   string `yaml:"version" json:"version"`
	Spec      any    `yaml:"spec" json:"spec"`
}

func envelope(r resource.Resource) resourceEnvelope {
	m := r.Metadata()
	return resourceEnvelope{
		Namespace: m.Namespace(),
		Type:      m.Type(),
		ID:        m.ID(),
		Version:   m.Version().String(),
		Spec:      r.Spec(),
	}
}

func formatResourcesYAML(items []resource.Resource) string {
	var sb strings.Builder
	enc := yaml.NewEncoder(&sb)
	enc.SetIndent(2)
	for _, r := range items {
		if err := enc.Encode(envelope(r)); err != nil {
			fmt.Fprintf(&sb, "# encode error: %v\n", err)
		}
	}
	_ = enc.Close()
	return sb.String()
}

func formatResourcesJSON(items []resource.Resource) string {
	envelopes := make([]resourceEnvelope, 0, len(items))
	for _, r := range items {
		envelopes = append(envelopes, envelope(r))
	}
	b, err := json.MarshalIndent(envelopes, "", "  ")
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return string(b)
}
