// +build rt_none

package agent

func (b *Agent) handleNetworkCmd(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "net", nil)
}
