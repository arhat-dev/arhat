// +build rt_none

package agent

func (b *Agent) handlePodCmd(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "pod", nil)
}
