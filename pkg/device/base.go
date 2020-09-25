// +build !nodev

package device

import (
	"context"
	"fmt"
	"sync"

	"arhat.dev/aranya-proto/aranyagopb"

	"arhat.dev/arhat/pkg/types"
)

type baseDevice struct {
	ctx context.Context

	connHashHex string
	conn        types.Connectivity

	state    aranyagopb.DeviceState
	stateMsg string

	mu *sync.RWMutex
}

func newBaseDevice(ctx context.Context, connHashHex string, conn types.Connectivity) *baseDevice {
	return &baseDevice{
		ctx: ctx,

		connHashHex: connHashHex,
		conn:        conn,

		state:    aranyagopb.DEVICE_STATE_CREATED,
		stateMsg: "Created",

		mu: new(sync.RWMutex),
	}
}

func (d *baseDevice) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	err := d.conn.Connect()
	if err != nil {
		d.state = aranyagopb.DEVICE_STATE_ERRORED
		d.stateMsg = err.Error()
		return fmt.Errorf("failed to connect device: %w", err)
	}

	d.state = aranyagopb.DEVICE_STATE_CONNECTED
	d.stateMsg = "Connected"

	return nil
}

func (d *baseDevice) Status() *aranyagopb.DeviceStatusMsg {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return aranyagopb.NewDeviceStatusMsg(aranyagopb.DEVICE_TYPE_NORMAL, d.connHashHex, d.state, d.stateMsg)
}
