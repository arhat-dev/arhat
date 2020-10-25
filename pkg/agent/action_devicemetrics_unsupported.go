// +build nodev nometrics nodevmetrics

package agent

func (b *Agent) handlePeripheralMetricsCollect(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "peripheral.metrics.collect", nil)
}
