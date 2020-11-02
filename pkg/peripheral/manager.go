package peripheral

import (
	"context"
	"fmt"
	"sync"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/arhat-proto/arhatgopb"
	"arhat.dev/libext/server"
	"arhat.dev/libext/util"
	"arhat.dev/pkg/log"
	"arhat.dev/pkg/wellknownerrors"

	"arhat.dev/arhat/pkg/conf"
)

func NewManager(ctx context.Context, config *conf.PeripheralExtensionConfig) *Manager {
	return &Manager{
		ctx: ctx,

		logger: log.Log.WithName("peripheral"),
		config: config,

		all:              make(map[string]*Conn),
		peripherals:      make(map[string]*Peripheral),
		metricsReporters: make(map[string]*MetricsReporter),

		metricsCache: NewMetricsCache(config.MetricsCacheTimeout),

		mu: new(sync.RWMutex),
	}
}

type Manager struct {
	ctx context.Context

	logger log.Interface
	config *conf.PeripheralExtensionConfig

	// key: name
	all map[string]*Conn
	// key: peripheral_id
	peripherals map[string]*Peripheral
	// key: name
	metricsReporters map[string]*MetricsReporter

	metricsCache *MetricsCache

	mu *sync.RWMutex

	extensions *sync.Map
}

func (m *Manager) CreateExtensionHandleFunc(
	extensionName string,
) (server.ExtensionHandleFunc, server.OutOfBandMsgHandleFunc) {
	handleFunc := func(c *server.ExtensionContext) {
		_, loaded := m.extensions.LoadOrStore(extensionName, c)
		if loaded {
			return
		}

		defer func() {
			c.Close()

			m.extensions.Delete(extensionName)
		}()

		select {
		case <-c.Context.Done():
			return
		case <-m.ctx.Done():
			return
		}
	}

	oobHandleFunc := func(msg *arhatgopb.Msg) {
		m.logger.I("received out of band message",
			log.String("extension", extensionName),
			log.Uint64("id", msg.Id),
			log.String("msg_type", msg.Kind.String()),
			log.Binary("payload", msg.Payload),
		)
	}

	return handleFunc, oobHandleFunc
}

func (m *Manager) connectTarget(
	extensionName string,
	target string,
	params map[string]string,
	tlsConfig *aranyagopb.TLSConfig,
) (_ *Conn, err error) {
	// TODO: determine peripheral id
	var id uint64 = 1

	v, ok := m.extensions.Load(extensionName)
	if !ok {
		return nil, fmt.Errorf("peripheral extension not found")
	}

	ec, ok := v.(*server.ExtensionContext)
	if !ok {
		return nil, fmt.Errorf("invalid non extension context stored")
	}

	connCmd := &arhatgopb.PeripheralConnectCmd{
		Target: target,
		Params: params,
		Tls:    nil,
	}
	if tlsConfig != nil {
		connCmd.Tls = &arhatgopb.TLSConfig{
			ServerName:         tlsConfig.ServerName,
			InsecureSkipVerify: tlsConfig.InsecureSkipVerify,
			MinVersion:         tlsConfig.MinVersion,
			MaxVersion:         tlsConfig.MaxVersion,
			CaCert:             tlsConfig.CaCert,
			Cert:               tlsConfig.Cert,
			Key:                tlsConfig.Key,
			CipherSuites:       tlsConfig.CipherSuites,
			NextProtos:         tlsConfig.NextProtos,
		}
	}

	cmd, err := util.NewCmd(
		ec.Codec.Marshal, arhatgopb.CMD_PERIPHERAL_CONNECT, id, 1, connCmd,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create peripheral connect cmd: %w", err)
	}

	resp, err := ec.SendCmd(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to send peripheral connect cmd: %w", err)
	}

	switch resp.Kind {
	case arhatgopb.MSG_DONE:
	case arhatgopb.MSG_ERROR:
	default:
		return nil, fmt.Errorf("unexpected")
	}

	return NewConnectivity(id, ec), nil
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

	conn, err := m.connectTarget(dc.Method, dc.Target, dc.Params, dc.Tls)
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
		dev := NewPeripheral(
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
		kind, name, found := func() (aranyagopb.PeripheralType, string, bool) {
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
			aranyagopb.NewPeripheralStatusMsg(kind, name, aranyagopb.PERIPHERAL_STATE_REMOVED, "Removed"),
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
