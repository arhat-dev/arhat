// +build nodev

package agent

func (b *Agent) handleDeviceList(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "device", nil)
}

func (b *Agent) handleDeviceEnsure(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "device", nil)
}

func (b *Agent) handleDeviceDelete(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "device", nil)
}

type deviceManager struct{}

func newDeviceManager() *deviceManager {
	return &deviceManager{}
}
