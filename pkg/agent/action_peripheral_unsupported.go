// +build noperipheral

package agent

import (
	"arhat.dev/arhat/pkg/conf"
)

type agentComponentPeripheral struct{}

func (b *Agent) createAndRegisterPeripheralExtensionManager(
	agentCtx context.Context,
	srv *server.Server,
	config *conf.PeripheralExtensionConfig,
) {
}

func (b *Agent) handlePeripheralList(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "peripheral.list", nil)
}

func (b *Agent) handlePeripheralEnsure(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "peripheral.ensure", nil)
}

func (b *Agent) handlePeripheralDelete(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "peripheral.delete", nil)
}

func (b *Agent) handlePeripheralOperation(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "peripheral.operate", nil)
}
