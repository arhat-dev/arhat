// +build rt_none

package agent

func (b *Agent) handleContainerNetworkList(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "net.ctr.list", nil)
}

func (b *Agent) handleContainerNetworkEnsure(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "net.ctr.ensure", nil)
}
