package tools

import (
	"github.com/mark3labs/mcp-go/server"

	"git.erwanleboucher.dev/eleboucher/talos-mcp/internal/talosclient"
)

// RegisterAll attaches every Talos MCP tool to the given server.
func RegisterAll(s *server.MCPServer, f *talosclient.Factory) {
	registerVersion(s, f)
	registerTime(s, f)
	registerHealth(s, f)
	registerMemory(s, f)
	registerProcesses(s, f)
	registerMounts(s, f)
	registerNetstat(s, f)
	registerDmesg(s, f)
	registerServiceLogs(s, f)
	registerServiceList(s, f)
	registerContainers(s, f)
	registerContainerStats(s, f)
	registerDisks(s, f)
	registerDiskStats(s, f)
	registerDiskUsage(s, f)
	registerList(s, f)
	registerRead(s, f)
	registerLogsContainers(s, f)
	registerImageList(s, f)
	registerKubeconfig(s, f)
	registerEtcdMembers(s, f)
	registerEtcdStatus(s, f)
	registerEtcdAlarms(s, f)
	registerGet(s, f)
}
