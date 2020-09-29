// +build nodev nometrics nodevmetrics

package agent

func (b *Agent) handleDeviceMetricsCollect(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "device.metrics.collect", nil)
}
