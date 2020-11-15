// +build !noextension
// +build !noextension_peripheral

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

package agent

import (
	"fmt"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/arhat-proto/arhatgopb"
	"arhat.dev/libext/server"

	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/arhat/pkg/peripheral"
)

type extensionComponentPeripheral struct {
	*peripheral.Manager
}

func (c *extensionComponentPeripheral) init(
	agent *Agent,
	srv *server.Server,
	config *conf.PeripheralExtensionConfig,
) {
	c.Manager = peripheral.NewManager(agent.ctx, config)
	srv.Handle(arhatgopb.EXTENSION_PERIPHERAL, c.Manager.CreateExtensionHandleFunc)
}

func (b *Agent) handlePeripheralList(sid uint64, data []byte) {
	cmd := new(aranyagopb.PeripheralListCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal PeripheralListCmd: %w", err))
		return
	}

	b.processInNewGoroutine(sid, "peripheral.list", func() {
		statusList := &aranyagopb.PeripheralStatusListMsg{
			Peripherals: b.Manager.GetAllStatuses(),
		}
		err = b.PostMsg(sid, aranyagopb.MSG_PERIPHERAL_STATUS_LIST, statusList)
		if err != nil {
			b.handleConnectivityError(sid, err)
			return
		}
	})
}

func (b *Agent) handlePeripheralEnsure(sid uint64, data []byte) {
	cmd := new(aranyagopb.PeripheralEnsureCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal PeripheralEnsureCmd: %w", err))
		return
	}

	b.processInNewGoroutine(sid, "peripheral.ensure", func() {
		err = b.Manager.Ensure(cmd)
		if err != nil {
			b.handleRuntimeError(sid, err)
			return
		}

		status := b.Manager.GetStatus(cmd.Name)
		err = b.PostMsg(sid, aranyagopb.MSG_PERIPHERAL_STATUS, status)
		if err != nil {
			b.handleConnectivityError(sid, err)
			return
		}
	})
}

func (b *Agent) handlePeripheralDelete(sid uint64, data []byte) {
	cmd := new(aranyagopb.PeripheralDeleteCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal PeripheralDeleteCmd: %w", err))
		return
	}

	if len(cmd.PeripheralNames) == 0 {
		b.handleRuntimeError(sid, errRequiredOptionsNotFound)
		return
	}

	b.processInNewGoroutine(sid, "peripheral.delete", func() {
		status := b.Manager.Delete(cmd.PeripheralNames...)
		if err != nil {
			b.handleRuntimeError(sid, err)
			return
		}

		err = b.PostMsg(
			sid,
			aranyagopb.MSG_PERIPHERAL_STATUS_LIST,
			&aranyagopb.PeripheralStatusListMsg{
				Peripherals: status,
			},
		)
		if err != nil {
			b.handleConnectivityError(sid, err)
			return
		}
	})
}

func (b *Agent) handlePeripheralOperate(sid uint64, data []byte) {
	cmd := new(aranyagopb.PeripheralOperateCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal PeripheralOperateCmd: %w", err))
		return
	}

	b.processInNewGoroutine(sid, "peripheral.operate", func() {
		var result [][]byte
		result, err = b.Manager.Operate(b.ctx, cmd.PeripheralName, cmd.OperationId, cmd.Data)
		if err != nil {
			b.handleRuntimeError(sid, err)
			return
		}

		err = b.PostMsg(
			sid,
			aranyagopb.MSG_PERIPHERAL_OPERATION_RESULT,
			&aranyagopb.PeripheralOperationResultMsg{
				Data: result,
			},
		)
		if err != nil {
			b.handleConnectivityError(sid, err)
			return
		}
	})
}
