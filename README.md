# talos-mcp

A Model Context Protocol server exposing read-only Talos Linux operations as
MCP tools. Stdio or Streamable HTTP transport.

## Tools

All accept optional `nodes`, `context`, and `endpoints` per call.

| Tool | `talosctl` equivalent |
| --- | --- |
| `talos_version` | `version` |
| `talos_time` | `time` |
| `talos_health` | `health` |
| `talos_memory` | `memory` |
| `talos_processes` | `processes` |
| `talos_mounts` | `mounts` |
| `talos_netstat` | `netstat` |
| `talos_dmesg` | `dmesg` (capped 1000 lines) |
| `talos_service_logs` | `logs <id>` (capped 1000 lines) |
| `talos_logs_containers` | `logs` |
| `talos_service_list` | `service` |
| `talos_containers` | `containers` |
| `talos_container_stats` | `stats` |
| `talos_image_list` | `image list` |
| `talos_disks` | `disks` |
| `talos_disk_stats` | `diskstats` |
| `talos_disk_usage` | `usage` |
| `talos_list` | `list <dir>` (capped 2000 entries) |
| `talos_read` | `read <path>` (capped 1 MiB) |
| `talos_kubeconfig` | `kubeconfig -` |
| `talos_etcd_members` | `etcd members` |
| `talos_etcd_status` | `etcd status` |
| `talos_etcd_alarms` | `etcd alarm list` |
| `talos_get` | `get <type> [<id>]` |

`talos_get` uses single-node proxying (COSI doesn't fan out).

## Transports

| Mode | Env |
| --- | --- |
| stdio (default) | `TALOS_MCP_TRANSPORT=stdio` |
| Streamable HTTP at `/mcp` | `TALOS_MCP_TRANSPORT=http`, `TALOS_MCP_HTTP_ADDR=:8080` |

## Configuration

| Env | Default |
| --- | --- |
| `TALOS_MCP_LOG_LEVEL` | `info` (`debug` also logs each call's arguments) |
| `TALOS_MCP_TOOL_TIMEOUT` | `20s` — `time.ParseDuration` format; clamps `talos_health.wait_timeout_seconds` |

`talos_disk_usage` caps results at 20000 entries.

## Auth

`client.WithDefaultConfig()` discovers talosconfig in order:

1. `$TALOSCONFIG`
2. `~/.talos/config`
3. `/var/run/secrets/talos.dev/config` — written by a `talos.dev/v1alpha1
   ServiceAccount` when the cluster grants
   `machine.features.kubernetesTalosAPIAccess` to the pod's namespace + role

`os:reader` is sufficient for every tool.

## Local

```
make build
make test
make smoke
```

## Docker

```
make docker
docker run --rm -p 8080:8080 \
  -e TALOS_MCP_TRANSPORT=http \
  -v ~/.talos/config:/talosconfig/config:ro \
  -e TALOSCONFIG=/talosconfig/config \
  talos-mcp:dev
```

## Kubernetes

`deploy/talos-mcp.yaml` declares a `talos.dev/v1alpha1 ServiceAccount` and a
`toolhive.stacklok.dev/v1alpha1 MCPServer`. Talos mints a Secret named
**exactly the same as the SA** (no `-talosconfig` suffix added), so the SA is
named `talos-mcp-talosconfig` and the pod mounts a Secret of the same name.

### Required Talos machineconfig

Every control-plane node must allow the pod's namespace + role. Add to your
machineconfig and `talosctl apply-config` to each control-plane node:

```yaml
machine:
  features:
    kubernetesTalosAPIAccess:
      enabled: true
      allowedKubernetesNamespaces:
        - <namespace-talos-mcp-runs-in>
      allowedRoles:
        - os:reader
```

No reboot needed — the change is runtime.

### Verify

```
kubectl -n <ns> get serviceaccount.talos.dev/talos-mcp-talosconfig -o jsonpath='{.status}'
kubectl -n <ns> get secret talos-mcp-talosconfig
```

`status` should be `{}` (no `failureReason`); the Secret should appear within
seconds. If `failureReason: 'Namespace is not allowed: …'`, the live
machineconfig hasn't been updated on at least one node.
