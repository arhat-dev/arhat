package peripheral

import (
	"context"
	"sync"

	"arhat.dev/aranya-proto/aranyagopb"
)

type basePeripheral struct {
	ctx context.Context

	name string
	conn *Conn

	state    aranyagopb.PeripheralState
	stateMsg string

	mu *sync.RWMutex
}

func newBasePeripheral(ctx context.Context, name string, conn *Conn) *basePeripheral {
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

	return &aranyagopb.PeripheralStatusMsg{
		Kind:    aranyagopb.PERIPHERAL_TYPE_NORMAL,
		Name:    d.name,
		State:   d.state,
		Message: d.stateMsg,
	}
}
