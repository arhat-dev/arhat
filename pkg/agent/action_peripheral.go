package agent

import (
	"context"
	"fmt"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/arhat-proto/arhatgopb"
	"arhat.dev/libext/server"

	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/arhat/pkg/peripheral"
)

type agentComponentPeripheral struct {
	peripheralManager *peripheral.Manager
}

func (b *agentComponentPeripheral) createAndRegisterPeripheralExtensionManager(
	agentCtx context.Context,
	srv *server.Server,
	config *conf.PeripheralExtensionConfig,
) {
	b.peripheralManager = peripheral.NewManager(agentCtx, config)
	srv.Handle(arhatgopb.EXTENSION_PERIPHERAL, b.peripheralManager.CreateExtensionHandleFunc)
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
			Peripherals: b.peripheralManager.GetAllStatuses(),
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
		err = b.peripheralManager.Ensure(cmd)
		if err != nil {
			b.handleRuntimeError(sid, err)
			return
		}

		status := b.peripheralManager.GetStatus(cmd.Name)
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
		status := b.peripheralManager.Delete(cmd.PeripheralNames...)
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

func (b *Agent) handlePeripheralOperation(sid uint64, data []byte) {
	cmd := new(aranyagopb.PeripheralOperateCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal PeripheralOperateCmd: %w", err))
		return
	}

	b.processInNewGoroutine(sid, "peripheral.operate", func() {
		var result [][]byte
		result, err = b.peripheralManager.Operate(b.Context(), cmd.PeripheralName, cmd.OperationId, cmd.Data)
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
