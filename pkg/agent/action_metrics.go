// +build !nometrics

/*
Copyright 2020 The arhat.dev Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package agent

import (
	"fmt"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/wellknownerrors"

	"arhat.dev/arhat/pkg/metrics"
)

func (b *Agent) handleMetricsConfig(sid uint64, data []byte) {
	cmd := new(aranyagopb.MetricsConfigCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal MetricsConfigCmd: %w", err))
		return
	}

	switch cmd.Target {
	case aranyagopb.METRICS_TARGET_NODE:
		b.processInNewGoroutine(sid, "metrics.config.node", func() {
			c, err := metrics.CreateNodeMetricsCollector(cmd)
			if err != nil {
				b.handleRuntimeError(sid, err)
				return
			}

			err = b.PostMsg(sid, aranyagopb.MSG_DONE, &aranyagopb.Empty{})
			if err != nil {
				b.handleConnectivityError(sid, err)
			} else {
				b.metricsMU.Lock()
				b.collectNodeMetrics = c
				b.metricsMU.Unlock()
			}
		})
	case aranyagopb.METRICS_TARGET_CONTAINER:
		b.processInNewGoroutine(sid, "metrics.config.container", func() {
			c, err := metrics.CreateContainerMetricsCollector(cmd)
			if err != nil {
				b.handleRuntimeError(sid, err)
				return
			}

			err = b.PostMsg(sid, aranyagopb.MSG_DONE, &aranyagopb.Empty{})
			if err != nil {
				b.handleConnectivityError(sid, err)
			} else {
				b.metricsMU.Lock()
				b.collectContainerMetrics = c
				b.metricsMU.Unlock()
			}
		})
	}
}

func (b *Agent) handleMetricsCollect(sid uint64, data []byte) {
	cmd := new(aranyagopb.MetricsCollectCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal MetricsCollectCmd: %w", err))
		return
	}

	switch cmd.Target {
	case aranyagopb.METRICS_TARGET_NODE:
		b.processInNewGoroutine(sid, "metrics.collect.node", func() {
			b.metricsMU.RLock()
			defer b.metricsMU.RUnlock()

			if b.collectNodeMetrics == nil {
				b.handleRuntimeError(sid, wellknownerrors.ErrNotSupported)
				return
			}

			metricsData, err := b.collectNodeMetrics()
			if err != nil {
				b.handleRuntimeError(sid, err)
			}

			_, err = b.PostData(sid, aranyagopb.MSG_DATA_METRICS, 0, true, metricsData)
			if err != nil {
				b.handleConnectivityError(sid, err)
			}
		})
	case aranyagopb.METRICS_TARGET_CONTAINER:
		b.processInNewGoroutine(sid, "metrics.collect.container", func() {
			b.metricsMU.RLock()
			defer b.metricsMU.RUnlock()

			if b.collectContainerMetrics == nil {
				b.handleRuntimeError(sid, wellknownerrors.ErrNotSupported)
				return
			}

			metricsData, err := b.collectContainerMetrics()
			if err != nil {
				b.handleRuntimeError(sid, err)
			}

			_, err = b.PostData(sid, aranyagopb.MSG_DATA_METRICS, 0, true, metricsData)
			if err != nil {
				b.handleConnectivityError(sid, err)
			}
		})
	default:
		b.handleUnknownCmd(sid, "metrics", cmd)
	}
}
