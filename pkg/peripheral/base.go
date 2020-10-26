// +build !noperipheral

package peripheral

import (
	"context"
	"sync"

	"arhat.dev/aranya-proto/aranyagopb"
)

type basePeripheral struct {
	ctx context.Context

	name string
	conn *Connectivity

	state    aranyagopb.PeripheralState
	stateMsg string

	mu *sync.RWMutex
}

func newBasePeripheral(ctx context.Context, name string, conn *Connectivity) *basePeripheral {
	return &basePeripheral{
		ctx: ctx,

		name: name,
		conn: conn,

		state:    aranyagopb.PERIPHERAL_STATE_CONNECTED,
		stateMsg: "Connected",

		mu: new(sync.RWMutex),
	}
}

func (d *basePeripheral) Status() *aranyagopb.PeripheralStatusMsg {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return aranyagopb.NewPeripheralStatusMsg(aranyagopb.PERIPHERAL_TYPE_NORMAL, d.name, d.state, d.stateMsg)
}
