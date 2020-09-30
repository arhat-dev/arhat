// +build !nodev

package device

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"

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

// MetricSpec defines how to collect one metric from device
type MetricSpec struct {
	Name                string
	ValueType           aranyagopb.DeviceMetric_ValueType
	ParamsForCollecting map[string]string

	ReportKey          MetricReportKey
	ParamsForReporting map[string]string
}

func NewDevice(
	ctx context.Context,
	connectorHashHex string,
	connector *Connectivity,
	operations []*aranyagopb.DeviceOperation,
	metrics []*aranyagopb.DeviceMetric,
) *Device {
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
				ParamsForCollecting: m.DeviceParams,

				ReportKey:          reportKey,
				ParamsForReporting: m.ReporterParams,
			})
		}
	}

	return &Device{
		baseDevice: newBaseDevice(ctx, connectorHashHex, connector),

		operations: ops,
		metrics:    ms,
	}
}

type Device struct {
	*baseDevice

	// operation_id -> params
	operations map[string]map[string]string
	metrics    []*MetricSpec
}

func (d *Device) Operate(ctx context.Context, id string, data []byte) ([][]byte, error) {
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

func (d *Device) Close() error {
	return d.conn.Close()
}

func convertMetricValue(value interface{}) (float64, error) {
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	case []byte:
		return strconv.ParseFloat(string(v), 64)
	default:
		return 0, fmt.Errorf("unsupported metrics value type")
	}
}

func NewMetricsReporter(
	ctx context.Context,
	connectorHashHex string,
	conn *Connectivity,
) *MetricsReporter {
	return &MetricsReporter{
		baseDevice: newBaseDevice(ctx, connectorHashHex, conn),
	}
}

type MetricsReporter struct {
	*baseDevice
}

func (r *MetricsReporter) Close() error {
	return r.conn.Close()
}
