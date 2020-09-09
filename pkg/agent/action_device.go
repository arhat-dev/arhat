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

func (b *Agent) handleDeviceCmd(sid uint64, data []byte) {
	cmd := new(aranyagopb.DeviceCmd)

	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal device cmd: %w", err))
		return
	}

	switch cmd.Action {
	case aranyagopb.LIST_DEVICES:
		b.processInNewGoroutine(sid, "device.list", func() {
			b.doDeviceList(sid)
		})
	case aranyagopb.ENSURE_DEVICE:
		spec := cmd.GetDeviceSpec()
		if spec == nil {
			b.handleRuntimeError(sid, errRequiredOptionsNotFound)
			return
		}

		b.processInNewGoroutine(sid, "device.ensure", func() {
			b.doDeviceEnsure(sid, spec)
		})
	case aranyagopb.REMOVE_DEVICE:
		id := cmd.GetDeviceId()
		if id == "" {
			b.handleRuntimeError(sid, errRequiredOptionsNotFound)
			return
		}

		b.processInNewGoroutine(sid, "device.remove", func() {
			b.doDeviceRemove(sid, id)
		})
	default:
		b.handleUnknownCmd(sid, "device", cmd)
	}
}

func (b *Agent) doDeviceList(sid uint64) {
	err := b.PostMsg(aranyagopb.NewDeviceStatusListMsg(sid, b.devices.getAllStatuses()))
	if err != nil {
		b.handleConnectivityError(sid, err)
		return
	}
}

func (b *Agent) doDeviceEnsure(sid uint64, spec *aranyagopb.Device) {
	err := b.devices.ensure(spec.Id, spec)
	if err != nil {
		b.handleRuntimeError(sid, err)
		return
	}

	deviceMsg := aranyagopb.NewDeviceStatusMsg(sid, b.devices.getStatus(spec.Id))
	if err := b.PostMsg(deviceMsg); err != nil {
		b.handleConnectivityError(sid, err)
		return
	}
}

func (b *Agent) doDeviceRemove(sid uint64, deviceID string) {
	status, err := b.devices.remove(deviceID)
	if err != nil {
		b.handleRuntimeError(sid, err)
		return
	}

	deviceMsg := aranyagopb.NewDeviceStatusMsg(sid, status)
	if err := b.PostMsg(deviceMsg); err != nil {
		b.handleConnectivityError(sid, err)
		return
	}
}

func newDevice(id string, conn, uploadConn types.DeviceConnectivity, operations []*aranyagopb.DeviceOperation, metrics []*aranyagopb.DeviceMetrics) *device {
	ops := make(map[string]*aranyagopb.DeviceOperation)
	for i, o := range operations {
		ops[o.Id] = operations[i]
	}

	ms := make(map[string]*aranyagopb.DeviceMetrics)
	for i, m := range metrics {
		ms[m.Id] = metrics[i]
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

	state    aranyagopb.DeviceStatus_State
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

func (d *device) status() *aranyagopb.DeviceStatus {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return aranyagopb.NewDeviceStatus(d.id, d.state, d.stateMsg)
}

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

func (m *deviceManager) ensure(id string, spec *aranyagopb.Device) error {
	err := func() error {
		m.mu.RLock()
		defer m.mu.RUnlock()

		if _, ok := m.devices[id]; ok {
			return wellknownerrors.ErrAlreadyExists
		}
		return nil
	}()
	if err != nil {
		return err
	}

	dc := spec.Connectivity
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
	if uc := spec.UploadConnectivity; uc != nil {
		newConn, ok := types.GetDeviceConnectivityFactory(uc.Transport, uc.Mode)
		if !ok {
			return fmt.Errorf("unsupported upload connectivity: %w", wellknownerrors.ErrNotSupported)
		}

		uploadConn, err = newConn(uc.Target, uc.Params, uc.Tls)
		if err != nil {
			return fmt.Errorf("failed to create upload connectivity: %w", err)
		}
	}

	d := newDevice(id, conn, uploadConn, spec.Operations, spec.Metrics)

	err = d.start()
	if err != nil {
		return fmt.Errorf("failed to start device %q: %w", id, err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.devices[id] = d

	return nil
}

func (m *deviceManager) remove(id string) (*aranyagopb.DeviceStatus, error) {
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

	return aranyagopb.NewDeviceStatus(id, aranyagopb.DEVICE_STATE_UNKNOWN, "removed"), nil
}

func (m *deviceManager) getStatus(id string) *aranyagopb.DeviceStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if d, ok := m.devices[id]; ok {
		return d.status()
	}

	return nil
}

func (m *deviceManager) getAllStatuses() []*aranyagopb.DeviceStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*aranyagopb.DeviceStatus
	for _, d := range m.devices {
		result = append(result, d.status())
	}

	return result
}

func (m *deviceManager) cleanup() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, d := range m.devices {
		// best effort
		_ = d.close()
	}
}
