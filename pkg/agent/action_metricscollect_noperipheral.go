// +build !nometrics
// +build noperipheral noperipheral_metrics

package agent

import (
	"fmt"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/wellknownerrors"

	"arhat.dev/arhat/pkg/types"
)

func (b *Agent) handleMetricsCollect(sid uint64, data []byte) {
	cmd := new(aranyagopb.MetricsCollectCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal MetricsCollectCmd: %w", err))
		return
	}

	type opTarget struct {
		name    string
		collect types.MetricsCollectFunc
	}

	b.metricsMU.RLock()

	op, ok := map[aranyagopb.MetricsTarget]*opTarget{
		aranyagopb.METRICS_TARGET_NODE:      {name: "node", collect: b.collectNodeMetrics},
		aranyagopb.METRICS_TARGET_CONTAINER: {name: "container", collect: b.collectContainerMetrics},
	}[cmd.Target]

	b.metricsMU.RUnlock()

	if !ok {
		b.handleUnknownCmd(sid, "metrics.collect", cmd)
		return
	}

	if op == nil {
		b.handleRuntimeError(sid, wellknownerrors.ErrNotSupported)
		return
	}

	b.processInNewGoroutine(sid, "metrics.collect."+op.name, func() {
		metrics, err := op.collect()
		if err != nil {
			b.handleRuntimeError(sid, err)
			return
		}

		data, err := b.encodeMetrics(metrics)
		if err != nil {
			b.handleConnectivityError(sid, fmt.Errorf("failed to encode metrics: %w", err))
			return
		}

		_, err = b.PostData(sid, aranyagopb.MSG_DATA_METRICS, 0, true, data)
		if err != nil {
			b.handleConnectivityError(sid, err)
			return
		}
	})
}
