// +build rt_none

package agent

func (b *Agent) handleContainerNetworkEnsure(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "ctr_net", nil)
}

func (b *Agent) handleImageEnsure(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "pod", nil)
}

func (b *Agent) handlePodList(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "pod", nil)
}

func (b *Agent) handlePodEnsure(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "pod", nil)
}

func (b *Agent) handlePodDelete(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "pod", nil)
}
