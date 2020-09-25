// +build !nodev

package device

import (
	"context"
	"fmt"
	"sync"
	"time"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/wellknownerrors"

	"arhat.dev/arhat/pkg/connectivity"
	"arhat.dev/arhat/pkg/types"
)

func NewManager(ctx context.Context, maxCacheTime time.Duration) *Manager {
	return &Manager{
		ctx: ctx,

		all:              make(map[string]types.Connectivity),
		devices:          make(map[string]*Device),
		metricsReporters: make(map[string]*MetricsReporter),

		metricsCache: NewMetricsCache(maxCacheTime),

		mu: new(sync.RWMutex),
	}
}

type Manager struct {
	ctx context.Context

	// key: connectivity_hash_hex
	all map[string]types.Connectivity

	// key: device_id
	devices map[string]*Device

	// key: connectivity_hash_hex
	metricsReporters map[string]*MetricsReporter

	metricsCache *MetricsCache

	mu *sync.RWMutex
}

func (m *Manager) Ensure(cmd *aranyagopb.DeviceEnsureCmd) (err error) {
	if cmd.ConnectorHashHex == "" {
		return fmt.Errorf("invalid empty connectivity hash hex")
	}

	switch cmd.Kind {
	case aranyagopb.DEVICE_TYPE_NORMAL:
		if cmd.DeviceId == "" {
			return fmt.Errorf("invalid empty device id for normal device")
		}
	case aranyagopb.DEVICE_TYPE_METRICS_REPORTER:
		if len(cmd.Operations) != 0 || len(cmd.Metrics) != 0 {
			return fmt.Errorf("metrics reporter should not have operations or collect metrics")
		}
	default:
		return fmt.Errorf("unknown device type: %v", cmd.Kind)
	}

	err = func() error {
		m.mu.RLock()
		defer m.mu.RUnlock()

		if _, ok := m.all[cmd.ConnectorHashHex]; ok {
			// TODO: ensure config up to date
			return wellknownerrors.ErrAlreadyExists
		}
		return nil
	}()
	if err != nil {
		return err
	}

	dc := cmd.Connector
	if dc == nil {
		return fmt.Errorf("required device connector spec not found")
	}

	createConn, ok := connectivity.NewConnectivity(dc.Method, dc.Mode)
	if !ok {
		return fmt.Errorf("unsupported device connectivity %q: %w", dc.Method, wellknownerrors.ErrNotSupported)
	}

	conn, err := createConn(dc.Target, dc.Params, dc.Tls)
	if err != nil {
		return fmt.Errorf("failed to create device connectivity: %w", err)
	}

	defer func() {
		if err != nil {
			_ = conn.Close()
		} else {
			m.mu.Lock()
			m.all[cmd.ConnectorHashHex] = conn
			m.mu.Unlock()
		}
	}()

	switch cmd.Kind {
	case aranyagopb.DEVICE_TYPE_NORMAL:
		dev := NewDevice(
			m.ctx, cmd.ConnectorHashHex, conn, cmd.Operations, cmd.Metrics,
		)

		err = dev.Start()
		if err != nil {
			_ = dev.Close()
			return fmt.Errorf("failed to start device %q: %w", cmd.DeviceId, err)
		}

		// nolint:unparam
		err = func() error {
			m.mu.Lock()
			defer m.mu.Unlock()

			m.devices[cmd.DeviceId] = dev
			return nil
		}()
	case aranyagopb.DEVICE_TYPE_METRICS_REPORTER:
		reporter := NewMetricsReporter(m.ctx, cmd.ConnectorHashHex, conn)

		err = reporter.Start()
		if err != nil {
			_ = reporter.Close()
			return fmt.Errorf("failed to start device %q: %w", cmd.DeviceId, err)
		}

		// nolint:unparam
		err = func() error {
			m.mu.Lock()
			defer m.mu.Unlock()

			m.metricsReporters[cmd.ConnectorHashHex] = reporter
			return nil
		}()
	}

	return
}

func (m *Manager) Delete(ids ...string) (result []*aranyagopb.DeviceStatusMsg) {
	for _, id := range ids {
		kind, hashHex, found := func() (aranyagopb.DeviceType, string, bool) {
			m.mu.RLock()
			defer m.mu.RUnlock()

			d, ok := m.devices[id]
			if ok {
				err := d.Close()
				if err != nil {
					_ = err
					// TODO: log error
				}

				return aranyagopb.DEVICE_TYPE_NORMAL, d.connHashHex, true
			}

			r, ok := m.metricsReporters[id]
			if ok {
				err := r.Close()
				if err != nil {
					_ = err
					// TODO: log error
				}

				return aranyagopb.DEVICE_TYPE_METRICS_REPORTER, r.connHashHex, true
			}

			return aranyagopb.DEVICE_TYPE_UNSPECIFIED, "", false
		}()

		if !found {
			result = append(result,
				aranyagopb.NewDeviceStatusMsg(kind, id, aranyagopb.DEVICE_STATE_UNKNOWN, "Not found"),
			)
			continue
		}

		m.mu.Lock()
		switch kind {
		case aranyagopb.DEVICE_TYPE_NORMAL:
			delete(m.devices, id)
		case aranyagopb.DEVICE_TYPE_METRICS_REPORTER:
			delete(m.metricsReporters, id)
		}
		m.mu.Unlock()

		result = append(result,
			aranyagopb.NewDeviceStatusMsg(kind, hashHex, aranyagopb.DEVICE_STATE_REMOVED, "Removed"),
		)
	}
	return
}

func (m *Manager) GetStatus(deviceID string) *aranyagopb.DeviceStatusMsg {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if d, ok := m.devices[deviceID]; ok {
		return d.Status()
	}

	return nil
}

func (m *Manager) Operate(deviceID, operationID string, data []byte) ([][]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	dev, ok := m.devices[deviceID]
	if !ok {
		return nil, wellknownerrors.ErrNotFound
	}

	return dev.Operate(operationID, data)
}

func (m *Manager) GetAllStatuses() []*aranyagopb.DeviceStatusMsg {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*aranyagopb.DeviceStatusMsg
	for _, d := range m.devices {
		result = append(result, d.Status())
	}

	return result
}

// nolint:unused
func (m *Manager) Cleanup() {
	ids := make(map[string]struct{})

	m.mu.RLock()
	for k, d := range m.devices {
		_ = d.Close()
		ids[k] = struct{}{}
	}

	for k, r := range m.metricsReporters {
		_ = r.Close()
		ids[k] = struct{}{}
	}

	for k, d := range m.all {
		_ = d.Close()
		ids[k] = struct{}{}
	}
	m.mu.RUnlock()

	m.mu.Lock()
	for k := range ids {
		delete(m.devices, k)
		delete(m.metricsReporters, k)
		delete(m.all, k)
	}
	m.mu.Unlock()

}
