// +build nodev

package agent

func (b *Agent) handleDeviceCmd(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "device", nil)
}

type deviceManager struct{}

func newDeviceManager() *deviceManager {
	return &deviceManager{}
}
