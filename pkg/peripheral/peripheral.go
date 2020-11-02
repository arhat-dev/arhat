package peripheral

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/wellknownerrors"
)

func hashStringMap(m map[string]string) string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := sha256.New()
	for _, k := range m {
		_, _ = h.Write([]byte(k))
		_, _ = h.Write([]byte(m[k]))
	}

	return hex.EncodeToString(h.Sum(nil))
}

type MetricReportKey struct {
	ReporterName  string
	ParamsHashHex string
}

// MetricSpec defines how to collect one metric from peripheral
type MetricSpec struct {
	Name                string
	ValueType           aranyagopb.PeripheralMetric_ValueType
	ParamsForCollecting map[string]string

	ReportKey          MetricReportKey
	ParamsForReporting map[string]string
}

func NewPeripheral(
	ctx context.Context,
	name string,
	connector *Conn,
	operations []*aranyagopb.PeripheralOperation,
	metrics []*aranyagopb.PeripheralMetric,
) *Peripheral {
	ops := make(map[string]map[string]string)
	for i, o := range operations {
		ops[o.OperationId] = operations[i].Params
	}

	var ms []*MetricSpec

	if len(metrics) > 0 {
		for _, m := range metrics {
			var reportKey MetricReportKey
			switch m.ReportMethod {
			case aranyagopb.REPORT_WITH_NODE_METRICS:
			case aranyagopb.REPORT_WITH_STANDALONE_CLIENT:
				reportKey.ReporterName = m.ReporterName
				fallthrough
			case aranyagopb.REPORT_WITH_ARHAT_CONNECTIVITY:
				reportKey.ParamsHashHex = hashStringMap(m.ReporterParams)
			}

			ms = append(ms, &MetricSpec{
				Name:                m.Name,
				ValueType:           m.ValueType,
				ParamsForCollecting: m.PeripheralParams,

				ReportKey:          reportKey,
				ParamsForReporting: m.ReporterParams,
			})
		}
	}

	return &Peripheral{
		basePeripheral: newBasePeripheral(ctx, name, connector),

		operations: ops,
		metrics:    ms,
	}
}

type Peripheral struct {
	*basePeripheral

	// operation_id -> params
	operations map[string]map[string]string
	metrics    []*MetricSpec
}

func (d *Peripheral) Operate(ctx context.Context, id string, data []byte) ([][]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	params, ok := d.operations[id]
	if !ok {
		return nil, wellknownerrors.ErrNotSupported
	}

	resp, err := d.conn.Operate(ctx, params, data)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (d *Peripheral) Close() error {
	return d.conn.Close()
}

func NewMetricsReporter(
	ctx context.Context,
	connectorHashHex string,
	conn *Conn,
) *MetricsReporter {
	return &MetricsReporter{
		basePeripheral: newBasePeripheral(ctx, connectorHashHex, conn),
	}
}

type MetricsReporter struct {
	*basePeripheral
}

func (r *MetricsReporter) Close() error {
	return r.conn.Close()
}
