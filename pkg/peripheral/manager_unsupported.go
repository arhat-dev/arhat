// +build nodev

package peripheral

import (
	"context"
)

func NewManager(ctx context.Context, _, _ interface{}) *Manager {
	return &Manager{}
}

type Manager struct{}

func (m *Manager) CacheMetrics(interface{})           {}
func (m *Manager) RetrieveCachedMetrics() interface{} { return nil }
