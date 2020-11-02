// +build noperipheral_metrics

package peripheral

import (
	"time"
)

func NewMetricsCache(maxCacheTime time.Duration) *MetricsCache {
	return &MetricsCache{}
}

type MetricsCache struct{}

func (m *MetricsCache) CacheMetrics(metrics interface{})   {}
func (m *MetricsCache) RetrieveCachedMetrics() interface{} { return nil }

func (m *Manager) CacheMetrics(interface{})           {}
func (m *Manager) RetrieveCachedMetrics() interface{} { return nil }
