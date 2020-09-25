// +build !nodev

package agent

import (
	"fmt"

	"arhat.dev/aranya-proto/aranyagopb"
)

func (b *Agent) handleDeviceList(sid uint64, data []byte) {
	cmd := new(aranyagopb.DeviceListCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal DeviceListCmd: %w", err))
		return
	}

	b.processInNewGoroutine(sid, "device.list", func() {
		statusList := aranyagopb.NewDeviceStatusListMsg(b.devices.GetAllStatuses())
		err = b.PostMsg(sid, aranyagopb.MSG_DEVICE_STATUS_LIST, statusList)
		if err != nil {
			b.handleConnectivityError(sid, err)
			return
		}
	})
}

func (b *Agent) handleDeviceEnsure(sid uint64, data []byte) {
	cmd := new(aranyagopb.DeviceEnsureCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal DeviceEnsureCmd: %w", err))
		return
	}

	b.processInNewGoroutine(sid, "device.ensure", func() {
		err = b.devices.Ensure(cmd)
		if err != nil {
			b.handleRuntimeError(sid, err)
			return
		}

		status := b.devices.GetStatus(cmd.DeviceId)
		err = b.PostMsg(sid, aranyagopb.MSG_DEVICE_STATUS, status)
		if err != nil {
			b.handleConnectivityError(sid, err)
			return
		}
	})
}

func (b *Agent) handleDeviceDelete(sid uint64, data []byte) {
	cmd := new(aranyagopb.DeviceDeleteCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal DeviceDeleteCmd: %w", err))
		return
	}

	if len(cmd.DeviceIds) == 0 {
		b.handleRuntimeError(sid, errRequiredOptionsNotFound)
		return
	}

	b.processInNewGoroutine(sid, "device.delete", func() {
		status := b.devices.Delete(append(cmd.DeviceIds, cmd.MetricsReporterHashHexes...)...)
		if err != nil {
			b.handleRuntimeError(sid, err)
			return
		}

		err = b.PostMsg(sid, aranyagopb.MSG_DEVICE_STATUS_LIST, aranyagopb.NewDeviceStatusListMsg(status))
		if err != nil {
			b.handleConnectivityError(sid, err)
			return
		}
	})
}

func (b *Agent) handleDeviceOperation(sid uint64, data []byte) {
	cmd := new(aranyagopb.DeviceOperateCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal DeviceOperateCmd: %w", err))
		return
	}

	b.processInNewGoroutine(sid, "device.operate", func() {
		var result [][]byte
		result, err = b.devices.Operate(cmd.DeviceId, cmd.OperationId, cmd.Data)
		if err != nil {
			b.handleRuntimeError(sid, err)
			return
		}

		err = b.PostMsg(sid, aranyagopb.MSG_DEVICE_OPERATION_RESULT, aranyagopb.NewDeviceOperationResultMsg(result))
		if err != nil {
			b.handleConnectivityError(sid, err)
			return
		}
	})
}
