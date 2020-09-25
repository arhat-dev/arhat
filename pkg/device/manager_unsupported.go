// +build nodev

package device

import (
	"context"
	"time"
)

func NewManager(ctx context.Context, maxCacheTime time.Duration) *Manager {
	return &Manager{}
}

type Manager struct{}

func (m *Manager) CacheMetrics(interface{})           {}
func (m *Manager) RetrieveCachedMetrics() interface{} { return nil }
