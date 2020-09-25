// +build nometrics

package types

type MetricsCollectFunc func() (interface{}, error)
