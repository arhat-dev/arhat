/*
Copyright 2020 The arhat.dev Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
