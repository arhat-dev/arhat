// +build !nodev
// +build !nometrics,!nodevmetrics

package agent

import (
	"fmt"

	"arhat.dev/aranya-proto/aranyagopb"
)

func (b *Agent) handleDeviceMetricsCollect(sid uint64, data []byte) {
	cmd := new(aranyagopb.DeviceMetricsCollectCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal DeviceMetricsCollectCmd: %w", err))
		return
	}

	b.processInNewGoroutine(sid, "device.metrics", func() {
		metricsForNode, paramsForAgent, metricsForAgent := b.devices.CollectMetrics(cmd.DeviceIds...)
		_, _ = paramsForAgent, metricsForAgent
		// TODO: add agent metrics report support

		b.devices.CacheMetrics(metricsForNode)
	})
}
