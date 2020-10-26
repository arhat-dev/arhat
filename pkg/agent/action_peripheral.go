// +build !noperipheral

package agent

import (
	"fmt"
	"net/http"

	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/arhat/pkg/peripheral"
	"arhat.dev/arhat/pkg/util/extensionutil"

	"arhat.dev/aranya-proto/aranyagopb"
)

func (b *Agent) createAndRegisterPeripheralExtensionManager(
	mux *http.ServeMux,
	config *conf.PeripheralExtensionConfig,
) {
	b.peripherals = peripheral.NewManager(b.ctx, config)
	mux.Handle("/peripherals", extensionutil.NewHandler(b.peripherals.Sync))
}

func (b *Agent) handlePeripheralList(sid uint64, data []byte) {
	cmd := new(aranyagopb.PeripheralListCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal PeripheralListCmd: %w", err))
		return
	}

	b.processInNewGoroutine(sid, "peripheral.list", func() {
		statusList := aranyagopb.NewPeripheralStatusListMsg(b.peripherals.GetAllStatuses())
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
		err = b.peripherals.Ensure(cmd)
		if err != nil {
			b.handleRuntimeError(sid, err)
			return
		}

		status := b.peripherals.GetStatus(cmd.Name)
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
		status := b.peripherals.Delete(cmd.PeripheralNames...)
		if err != nil {
			b.handleRuntimeError(sid, err)
			return
		}

		err = b.PostMsg(sid, aranyagopb.MSG_PERIPHERAL_STATUS_LIST, aranyagopb.NewPeripheralStatusListMsg(status))
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
		result, err = b.peripherals.Operate(b.Context(), cmd.PeripheralName, cmd.OperationId, cmd.Data)
		if err != nil {
			b.handleRuntimeError(sid, err)
			return
		}

		err = b.PostMsg(
			sid,
			aranyagopb.MSG_PERIPHERAL_OPERATION_RESULT,
			aranyagopb.NewPeripheralOperationResultMsg(result),
		)
		if err != nil {
			b.handleConnectivityError(sid, err)
			return
		}
	})
}
