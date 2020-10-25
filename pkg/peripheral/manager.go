// +build !nodev

package peripheral

import (
	"context"
	"fmt"
	"sync"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/arhat-proto/arhatgopb"
	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/pkg/log"
	"arhat.dev/pkg/wellknownerrors"
)

type FactoryFunc func(
	ctx context.Context,
	target string,
	params map[string]string,
	tlsConfig *aranyagopb.TLSConfig,
) (*Connectivity, error)

// RegisterConnectivity add one connectivity method for peripheral
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

func NewManager(ctx context.Context, config *conf.PeripheralExtensionConfig) *Manager {
	return &Manager{
		ctx: ctx,

		logger: log.Log.WithName("peripheral"),
		config: config,

		all:              make(map[string]*Connectivity),
		peripherals:      make(map[string]*Device),
		metricsReporters: make(map[string]*MetricsReporter),

		metricsCache: NewMetricsCache(config.MaxMetricsCacheTime),

		mu: new(sync.RWMutex),

		supportedConnectivity: make(map[string]FactoryFunc),
	}
}

type Manager struct {
	ctx context.Context

	logger log.Interface
	config *conf.PeripheralExtensionConfig

	// key: connectivity_hash_hex
	all map[string]*Connectivity
	// key: peripheral_id
	peripherals map[string]*Device
	// key: connectivity_hash_hex
	metricsReporters map[string]*MetricsReporter

	metricsCache *MetricsCache

	mu *sync.RWMutex

	supportedConnectivity map[string]FactoryFunc
}

func (m *Manager) Sync(stopSig <-chan struct{}, msgCh <-chan *arhatgopb.Msg, cmdCh chan<- *arhatgopb.Cmd) error {
	var (
		errCh = make(chan error)
	)

	sendErr := func(err error) {
		select {
		case <-stopSig:
		case <-m.ctx.Done():
		case errCh <- err:
		}
	}

	sendCmd := func(cmd *arhatgopb.Cmd) {
		m.logger.V("sending cmd")
		select {
		case <-stopSig:
		case <-m.ctx.Done():
		case cmdCh <- cmd:
		}
	}

	go func() {
		started := false

		peripheralID := uint64(0)
		peripheralMU := new(sync.RWMutex)

		nextDeviceID := func() uint64 {
			peripheralMU.Lock()
			defer peripheralMU.Unlock()
			peripheralID++
			return peripheralID
		}
		connectedDevices := make(map[uint64]chan *arhatgopb.Msg)

		m.logger.I("receiving msgs")
		for msg := range msgCh {
			if started {
				switch msg.Kind {
				case arhatgopb.MSG_REGISTER:
					sendErr(fmt.Errorf("unexpected multiple register message"))
					return
				default:
					peripheralMU.RLock()
					ch, ok := connectedDevices[msg.Id]
					peripheralMU.RUnlock()
					if !ok {
						// TODO: log error
						continue
					}

					// TODO: use goroutine pool
					go func() {
						select {
						case <-stopSig:
						case <-m.ctx.Done():
						case ch <- msg:
						}
					}()
				}

				continue
			}

			if msg.Kind != arhatgopb.MSG_REGISTER {
				sendErr(fmt.Errorf("expecting register as first message"))
				return
			}

			r := new(arhatgopb.RegisterMsg)
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
					peripheralID := nextDeviceID()
					var cmd *arhatgopb.Cmd
					cmd, err = arhatgopb.NewCmd(peripheralID, 0, &arhatgopb.PeripheralConnectCmd{
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
						return nil, fmt.Errorf("failed to create peripheral cmd: %w", err)
					}

					msgCh := make(chan *arhatgopb.Msg)
					peripheralMU.Lock()
					connectedDevices[peripheralID] = msgCh
					peripheralMU.Unlock()

					sendCmd(cmd)

					return NewConnectivity(stopSig, peripheralID, cmdCh, msgCh, func() {
						peripheralMU.Lock()
						delete(connectedDevices, peripheralID)
						peripheralMU.Unlock()
					}), nil
				},
			)

			if err != nil {
				// TODO: log error, do not send error
			}
		}
	}()

	select {
	case <-stopSig:
		return nil
	case <-m.ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

func (m *Manager) Ensure(cmd *aranyagopb.PeripheralEnsureCmd) (err error) {
	if cmd.Name == "" {
		return fmt.Errorf("invalid empty name")
	}

	switch cmd.Kind {
	case aranyagopb.PERIPHERAL_TYPE_NORMAL:
		if cmd.Name == "" {
			return fmt.Errorf("invalid empty peripheral id for normal peripheral")
		}
	case aranyagopb.PERIPHERAL_TYPE_METRICS_REPORTER:
		if len(cmd.Operations) != 0 || len(cmd.Metrics) != 0 {
			return fmt.Errorf("metrics reporter should not have operations or collect metrics")
		}
	default:
		return fmt.Errorf("unknown peripheral type: %v", cmd.Kind)
	}

	err = func() error {
		m.mu.RLock()
		defer m.mu.RUnlock()

		if _, ok := m.all[cmd.Name]; ok {
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
		return fmt.Errorf("required peripheral connector spec not found")
	}

	conn, err := m.NewConnectivity(dc.Method, dc.Target, dc.Params, dc.Tls)
	if err != nil {
		return fmt.Errorf("failed to create peripheral connectivity: %w", err)
	}

	defer func() {
		if err != nil {
			_ = conn.Close()
		} else {
			m.mu.Lock()
			m.all[cmd.Name] = conn
			m.mu.Unlock()
		}
	}()

	switch cmd.Kind {
	case aranyagopb.PERIPHERAL_TYPE_NORMAL:
		dev := NewDevice(
			m.ctx, cmd.Name, conn, cmd.Operations, cmd.Metrics,
		)

		// nolint:unparam
		err = func() error {
			m.mu.Lock()
			defer m.mu.Unlock()

			m.peripherals[cmd.Name] = dev
			return nil
		}()
	case aranyagopb.PERIPHERAL_TYPE_METRICS_REPORTER:
		reporter := NewMetricsReporter(m.ctx, cmd.Name, conn)

		// nolint:unparam
		err = func() error {
			m.mu.Lock()
			defer m.mu.Unlock()

			m.metricsReporters[cmd.Name] = reporter
			return nil
		}()
	}

	return
}

func (m *Manager) Delete(ids ...string) (result []*aranyagopb.PeripheralStatusMsg) {
	for _, id := range ids {
		kind, hashHex, found := func() (aranyagopb.PeripheralType, string, bool) {
			m.mu.RLock()
			defer m.mu.RUnlock()

			d, ok := m.peripherals[id]
			if ok {
				err := d.Close()
				if err != nil {
					_ = err
					// TODO: log error
				}

				return aranyagopb.PERIPHERAL_TYPE_NORMAL, d.name, true
			}

			r, ok := m.metricsReporters[id]
			if ok {
				err := r.Close()
				if err != nil {
					_ = err
					// TODO: log error
				}

				return aranyagopb.PERIPHERAL_TYPE_METRICS_REPORTER, r.name, true
			}

			return 0, "", false
		}()

		if !found {
			result = append(result,
				aranyagopb.NewPeripheralStatusMsg(kind, id, aranyagopb.PERIPHERAL_STATE_UNKNOWN, "Not found"),
			)
			continue
		}

		m.mu.Lock()
		switch kind {
		case aranyagopb.PERIPHERAL_TYPE_NORMAL:
			delete(m.peripherals, id)
		case aranyagopb.PERIPHERAL_TYPE_METRICS_REPORTER:
			delete(m.metricsReporters, id)
		}
		m.mu.Unlock()

		result = append(result,
			aranyagopb.NewPeripheralStatusMsg(kind, hashHex, aranyagopb.PERIPHERAL_STATE_REMOVED, "Removed"),
		)
	}
	return
}

func (m *Manager) GetStatus(peripheralID string) *aranyagopb.PeripheralStatusMsg {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if d, ok := m.peripherals[peripheralID]; ok {
		return d.Status()
	}

	return nil
}

func (m *Manager) Operate(ctx context.Context, peripheralID, operationID string, data []byte) ([][]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	dev, ok := m.peripherals[peripheralID]
	if !ok {
		return nil, wellknownerrors.ErrNotFound
	}

	return dev.Operate(ctx, operationID, data)
}

func (m *Manager) GetAllStatuses() []*aranyagopb.PeripheralStatusMsg {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*aranyagopb.PeripheralStatusMsg
	for _, d := range m.peripherals {
		result = append(result, d.Status())
	}

	return result
}

// nolint:unused
func (m *Manager) Cleanup() {
	ids := make(map[string]struct{})

	m.mu.RLock()
	for k, d := range m.peripherals {
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
		delete(m.peripherals, k)
		delete(m.metricsReporters, k)
		delete(m.all, k)
	}
	m.mu.Unlock()

}
