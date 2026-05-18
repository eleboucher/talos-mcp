// Each Talos resource package registers its protobuf type with COSI in
// init(). Without these blank imports, COSI calls silently return zero items
// because the spec can't be deserialized.

package tools

import (
	_ "github.com/siderolabs/talos/pkg/machinery/resources/block"
	_ "github.com/siderolabs/talos/pkg/machinery/resources/cluster"
	_ "github.com/siderolabs/talos/pkg/machinery/resources/config"
	_ "github.com/siderolabs/talos/pkg/machinery/resources/cri"
	_ "github.com/siderolabs/talos/pkg/machinery/resources/etcd"
	_ "github.com/siderolabs/talos/pkg/machinery/resources/files"
	_ "github.com/siderolabs/talos/pkg/machinery/resources/hardware"
	_ "github.com/siderolabs/talos/pkg/machinery/resources/k8s"
	_ "github.com/siderolabs/talos/pkg/machinery/resources/kubeaccess"
	_ "github.com/siderolabs/talos/pkg/machinery/resources/kubespan"
	_ "github.com/siderolabs/talos/pkg/machinery/resources/network"
	_ "github.com/siderolabs/talos/pkg/machinery/resources/perf"
	_ "github.com/siderolabs/talos/pkg/machinery/resources/runtime"
	_ "github.com/siderolabs/talos/pkg/machinery/resources/secrets"
	_ "github.com/siderolabs/talos/pkg/machinery/resources/siderolink"
	_ "github.com/siderolabs/talos/pkg/machinery/resources/time"
	_ "github.com/siderolabs/talos/pkg/machinery/resources/v1alpha1"
)
