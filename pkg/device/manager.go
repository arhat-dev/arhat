// +build !nodev

package device

import (
	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/arhat-proto/arhatgopb"
	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/pkg/wellknownerrors"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"sync"
)

type FactoryFunc func(
	ctx context.Context,
	target string,
	params map[string]string,
	tlsConfig *aranyagopb.TLSConfig,
) (*Connectivity, error)

// RegisterConnectivity add one connectivity method for device
func (m *Manager) RegisterConnectivity(
	name string,
	factory FactoryFunc,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.supportedConnectivity[name]; ok {
		return fmt.Errorf("connectivity %s already registered", name)
	}

	m.supportedConnectivity[name] = factory

	return nil
}

func (m *Manager) NewConnectivity(
	name string,
	target string,
	params map[string]string,
	tlsConfig *aranyagopb.TLSConfig,
) (*Connectivity, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	factory, ok := m.supportedConnectivity[name]

	if !ok {
		return nil, fmt.Errorf("unsupported connectivity")
	}

	return factory(m.ctx, target, params, tlsConfig)
}

func NewManager(ctx context.Context, config *conf.DeviceExtensionConfig, server *grpc.Server) *Manager {

	mgr := &Manager{
		ctx: ctx,

		config: config,

		all:              make(map[string]*Connectivity),
		devices:          make(map[string]*Device),
		metricsReporters: make(map[string]*MetricsReporter),

		metricsCache: NewMetricsCache(config.MaxMetricsCacheTime),

		mu: new(sync.RWMutex),

		supportedConnectivity: make(map[string]FactoryFunc),
	}

	arhatgopb.RegisterDeviceExtensionServer(server, mgr)

	return mgr
}

type Manager struct {
	ctx context.Context

	config *conf.DeviceExtensionConfig

	// key: connectivity_hash_hex
	all map[string]*Connectivity
	// key: device_id
	devices map[string]*Device
	// key: connectivity_hash_hex
	metricsReporters map[string]*MetricsReporter

	metricsCache *MetricsCache

	mu *sync.RWMutex

	supportedConnectivity map[string]FactoryFunc
}

func (m *Manager) Sync(srv arhatgopb.DeviceExtension_SyncServer) error {
	var (
		errCh = make(chan error)
	)

	sendErr := func(err error) {
		select {
		case <-srv.Context().Done():
		case <-m.ctx.Done():
		case errCh <- err:
		}
	}

	go func() {
		started := false

		deviceID := uint64(0)
		deviceMU := new(sync.RWMutex)

		nextDeviceID := func() uint64 {
			deviceMU.Lock()
			defer deviceMU.Unlock()
			deviceID++
			return deviceID
		}
		connectedDevices := make(map[uint64]chan *arhatgopb.DeviceMsg)
		for {
			msg, err := srv.Recv()
			if err != nil {
				return
			}

			if !started {
				if msg.Kind != arhatgopb.MSG_DEV_REGISTER {
					sendErr(fmt.Errorf("expecting register as first message"))
					return
				}

				r := new(arhatgopb.DeviceRegisterMsg)
				err := r.Unmarshal(msg.Payload)
				if err != nil {
					sendErr(fmt.Errorf("failed to unmarhsal very first register message: %w", err))
					return
				}

				started = true

				err = m.RegisterConnectivity(r.Name,
					func(
						ctx context.Context,
						target string,
						params map[string]string,
						tlsConfig *aranyagopb.TLSConfig,
					) (*Connectivity, error) {
						deviceID := nextDeviceID()
						cmd, err := arhatgopb.NewDeviceCmd(deviceID, 0, &arhatgopb.DeviceConnectCmd{
							Target: target,
							Params: params,
							Tls: &arhatgopb.TLSConfig{
								ServerName:         tlsConfig.ServerName,
								InsecureSkipVerify: tlsConfig.InsecureSkipVerify,
								MinVersion:         tlsConfig.MinVersion,
								MaxVersion:         tlsConfig.MaxVersion,
								CaCert:             tlsConfig.CaCert,
								Cert:               tlsConfig.Cert,
								Key:                tlsConfig.Key,
								CipherSuites:       tlsConfig.CipherSuites,
								NextProtos:         tlsConfig.NextProtos,
							},
						})
						if err != nil {
							return nil, fmt.Errorf("failed to create device cmd: %w", err)
						}

						err = srv.Send(cmd)
						if err != nil {
							return nil, err
						}

						msgCh := make(chan *arhatgopb.DeviceMsg)
						deviceMU.Lock()
						connectedDevices[deviceID] = msgCh
						deviceMU.Unlock()

						return NewConnectivity(m.ctx, deviceID, srv, msgCh, func() {
							deviceMU.Lock()
							delete(connectedDevices, deviceID)
							deviceMU.Unlock()
						}), nil
					},
				)

				if err != nil {
					sendErr(err)
					return
				}

				continue
			}

			switch msg.Kind {
			case arhatgopb.MSG_DEV_REGISTER:
				sendErr(fmt.Errorf("unexpected multiple register message"))
				return
			default:
				deviceMU.RLock()
				ch, ok := connectedDevices[msg.DeviceId]
				deviceMU.RUnlock()
				if !ok {
					// TODO: log error
					continue
				}

				// TODO: use goroutine pool
				go func() {
					select {
					case <-srv.Context().Done():
					case <-m.ctx.Done():
					case ch <- msg:
					}
				}()
			}
		}
	}()

	select {
	case <-srv.Context().Done():
		return nil
	case <-m.ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
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

	conn, err := m.NewConnectivity(dc.Method, dc.Target, dc.Params, dc.Tls)
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

		// nolint:unparam
		err = func() error {
			m.mu.Lock()
			defer m.mu.Unlock()

			m.devices[cmd.DeviceId] = dev
			return nil
		}()
	case aranyagopb.DEVICE_TYPE_METRICS_REPORTER:
		reporter := NewMetricsReporter(m.ctx, cmd.ConnectorHashHex, conn)

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

func (m *Manager) Operate(ctx context.Context, deviceID, operationID string, data []byte) ([][]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	dev, ok := m.devices[deviceID]
	if !ok {
		return nil, wellknownerrors.ErrNotFound
	}

	return dev.Operate(ctx, operationID, data)
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
