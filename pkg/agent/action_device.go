// +build !nodev

package agent

import (
	"bytes"
	"fmt"
	"sync"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/wellknownerrors"

	"arhat.dev/arhat/pkg/types"
)

func (b *Agent) handleDeviceList(sid uint64, data []byte) {
	cmd := new(aranyagopb.DeviceListCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal DeviceListCmd: %w", err))
		return
	}

	b.processInNewGoroutine(sid, "device.list", func() {
		statusList := aranyagopb.NewDeviceStatusListMsg(b.devices.getAllStatuses())
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
		err = b.devices.ensure(cmd)
		if err != nil {
			b.handleRuntimeError(sid, err)
			return
		}

		status := b.devices.getStatus(cmd.DeviceId)
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

	for _, id := range cmd.DeviceIds {
		deviceID := id
		b.processInNewGoroutine(sid, "device.remove", func() {
			var status *aranyagopb.DeviceStatusMsg
			status, err := b.devices.remove(deviceID)
			if err != nil {
				b.handleRuntimeError(sid, err)
				return
			}

			err = b.PostMsg(sid, aranyagopb.MSG_DEVICE_STATUS, status)
			if err != nil {
				b.handleConnectivityError(sid, err)
				return
			}
		})
	}
}

func newDevice(
	id string,
	conn, uploadConn types.DeviceConnectivity,
	operations []*aranyagopb.DeviceOperation,
	metrics []*aranyagopb.DeviceMetrics,
) *device {
	ops := make(map[string]*aranyagopb.DeviceOperation)
	for i, o := range operations {
		ops[o.Id] = operations[i]
	}

	ms := make(map[string]*aranyagopb.DeviceMetrics)
	for i, m := range metrics {
		ms[m.Name] = metrics[i]
	}

	return &device{
		id: id,

		conn:       conn,
		uploadConn: uploadConn,
		operations: ops,
		metrics:    ms,

		state:    aranyagopb.DEVICE_STATE_UNKNOWN,
		stateMsg: "Unknown",
		mu:       new(sync.RWMutex),
	}
}

type device struct {
	id string

	conn       types.DeviceConnectivity
	uploadConn types.DeviceConnectivity
	operations map[string]*aranyagopb.DeviceOperation
	metrics    map[string]*aranyagopb.DeviceMetrics

	state    aranyagopb.DeviceStatusMsg_State
	stateMsg string

	mu *sync.RWMutex
}

func (d *device) start() error {
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

func (d *device) status() *aranyagopb.DeviceStatusMsg {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return aranyagopb.NewDeviceStatusMsg(d.id, d.state, d.stateMsg)
}

// nolint:unused
func (d *device) operate(id string) ([][]byte, error) {
	op, ok := d.operations[id]
	if !ok {
		return nil, wellknownerrors.ErrNotFound
	}

	resp, err := d.conn.Operate(op.TransportParams)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// nolint:unused
func (d *device) collectMetrics() ([]byte, error) {
	buf := new(bytes.Buffer)
	for _, m := range d.metrics {
		metricsBytes, err := d.conn.CollectMetrics(m.TransportParams)
		if err != nil {
			return nil, err
		}

		_ = metricsBytes
	}

	return buf.Bytes(), nil
}

func (d *device) close() error {
	return d.conn.Close()
}

func newDeviceManager() *deviceManager {
	return &deviceManager{
		devices: make(map[string]*device),
		mu:      new(sync.RWMutex),
	}
}

type deviceManager struct {
	devices map[string]*device
	mu      *sync.RWMutex
}

func (m *deviceManager) ensure(cmd *aranyagopb.DeviceEnsureCmd) error {
	err := func() error {
		m.mu.RLock()
		defer m.mu.RUnlock()

		if _, ok := m.devices[cmd.DeviceId]; ok {
			return wellknownerrors.ErrAlreadyExists
		}
		return nil
	}()
	if err != nil {
		return err
	}

	dc := cmd.DeviceConnectivity
	if dc == nil {
		return errRequiredOptionsNotFound
	}

	newConn, ok := types.GetDeviceConnectivityFactory(dc.Transport, dc.Mode)
	if !ok {
		return fmt.Errorf("unsupported device connectivity %q: %w", dc.Transport, wellknownerrors.ErrNotSupported)
	}

	conn, err := newConn(dc.Target, dc.Params, dc.Tls)
	if err != nil {
		return fmt.Errorf("failed to create device connectivity: %w", err)
	}

	var uploadConn types.DeviceConnectivity
	if uc := cmd.UploadConnectivity; uc != nil {
		newConn, ok := types.GetDeviceConnectivityFactory(uc.Transport, uc.Mode)
		if !ok {
			return fmt.Errorf("unsupported upload connectivity: %w", wellknownerrors.ErrNotSupported)
		}

		uploadConn, err = newConn(uc.Target, uc.Params, uc.Tls)
		if err != nil {
			return fmt.Errorf("failed to create upload connectivity: %w", err)
		}
	}

	d := newDevice(cmd.DeviceId, conn, uploadConn, cmd.DeviceOperations, cmd.DeviceMetrics)

	err = d.start()
	if err != nil {
		return fmt.Errorf("failed to start device %q: %w", cmd.DeviceId, err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.devices[cmd.DeviceId] = d

	return nil
}

func (m *deviceManager) remove(id string) (*aranyagopb.DeviceStatusMsg, error) {
	err := func() error {
		m.mu.RLock()
		defer m.mu.RUnlock()

		d, ok := m.devices[id]
		if !ok {
			return wellknownerrors.ErrNotFound
		}

		err := d.close()
		if err != nil {
			return err
		}

		return nil
	}()

	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.devices, id)

	return aranyagopb.NewDeviceStatusMsg(id, aranyagopb.DEVICE_STATE_UNKNOWN, "removed"), nil
}

func (m *deviceManager) getStatus(id string) *aranyagopb.DeviceStatusMsg {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if d, ok := m.devices[id]; ok {
		return d.status()
	}

	return nil
}

func (m *deviceManager) getAllStatuses() []*aranyagopb.DeviceStatusMsg {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*aranyagopb.DeviceStatusMsg
	for _, d := range m.devices {
		result = append(result, d.status())
	}

	return result
}

// nolint:unused
func (m *deviceManager) cleanup() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, d := range m.devices {
		// best effort
		_ = d.close()
	}
}
