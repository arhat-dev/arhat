// +build !nodev

package device

import (
	"context"
	"sync"

	"arhat.dev/aranya-proto/aranyagopb"
)

type baseDevice struct {
	ctx context.Context

	connHashHex string
	conn        *Connectivity

	state    aranyagopb.DeviceState
	stateMsg string

	mu *sync.RWMutex
}

func newBaseDevice(ctx context.Context, connHashHex string, conn *Connectivity) *baseDevice {
	return &baseDevice{
		ctx: ctx,

		connHashHex: connHashHex,
		conn:        conn,

		state:    aranyagopb.DEVICE_STATE_CONNECTED,
		stateMsg: "Connected",

		mu: new(sync.RWMutex),
	}
}

func (d *baseDevice) Status() *aranyagopb.DeviceStatusMsg {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return aranyagopb.NewDeviceStatusMsg(aranyagopb.DEVICE_TYPE_NORMAL, d.connHashHex, d.state, d.stateMsg)
}
