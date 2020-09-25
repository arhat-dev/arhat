package types

import "time"

type DeviceMetricValue struct {
	// Value of the this metric
	Value interface{}

	// Timestamp in unix millisecond
	Timestamp *time.Time
}

type Connectivity interface {
	Connect() error

	// Operate the device via established connection
	Operate(params map[string]string, data []byte) ([][]byte, error)

	// CollectMetrics collects all existing metrics for one metric kind
	CollectMetrics(params map[string]string) ([]*DeviceMetricValue, error)

	Close() error
}
