// +build nometrics

package agent

func (b *Agent) handleMetricsConfig(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "metrics", nil)
}

func (b *Agent) handleMetricsCollect(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "metrics", nil)
}
