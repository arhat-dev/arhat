// +build !nodev
// +build !nometrics,!nodevmetrics

package agent

import (
	"fmt"

	"arhat.dev/aranya-proto/aranyagopb"
)

func (b *Agent) handlePeripheralMetricsCollect(sid uint64, data []byte) {
	cmd := new(aranyagopb.PeripheralMetricsCollectCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal PeripheralMetricsCollectCmd: %w", err))
		return
	}

	b.processInNewGoroutine(sid, "peripheral.metrics", func() {
		metricsForNode, paramsForAgent, metricsForAgent := b.peripherals.CollectMetrics(cmd.PeripheralNames...)
		_, _ = paramsForAgent, metricsForAgent
		// TODO: add agent metrics report support

		b.peripherals.CacheMetrics(metricsForNode)
	})
}
