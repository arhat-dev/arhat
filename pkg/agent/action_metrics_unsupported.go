// +build nometrics

package agent

func (b *Agent) handleMetricsCmd(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "metrics", nil)
}
