// +build nodev

package agent

func (b *Agent) handleDeviceList(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "device.list", nil)
}

func (b *Agent) handleDeviceEnsure(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "device.ensure", nil)
}

func (b *Agent) handleDeviceDelete(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "device.delete", nil)
}

func (b *Agent) handleDeviceOperation(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "device.operate", nil)
}
