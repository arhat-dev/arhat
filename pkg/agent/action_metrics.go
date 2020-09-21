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
	"arhat.dev/arhat/pkg/types"
)

func (b *Agent) handleMetricsConfig(sid uint64, data []byte) {
	cmd := new(aranyagopb.MetricsConfigCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal MetricsConfigCmd: %w", err))
		return
	}

	type opTarget struct {
		name   string
		ptr    *types.MetricsCollectFunc
		create func(*aranyagopb.MetricsConfigCmd) (types.MetricsCollectFunc, error)
	}

	op, ok := map[aranyagopb.MetricsTarget]*opTarget{
		aranyagopb.METRICS_TARGET_NODE: {
			name: "node", ptr: &b.collectNodeMetrics, create: metrics.CreateNodeMetricsCollector,
		},
		aranyagopb.METRICS_TARGET_CONTAINER: {
			name: "container", ptr: &b.collectContainerMetrics, create: metrics.CreateContainerMetricsCollector,
		},
	}[cmd.Target]

	if !ok {
		b.handleUnknownCmd(sid, "metrics.config", cmd)
		return
	}

	if op == nil {
		b.handleRuntimeError(sid, wellknownerrors.ErrNotSupported)
		return
	}

	b.processInNewGoroutine(sid, "metrics.config."+op.name, func() {
		c, err := op.create(cmd)
		if err != nil {
			b.handleRuntimeError(sid, err)
			return
		}

		err = b.PostMsg(sid, aranyagopb.MSG_DONE, &aranyagopb.Empty{})
		if err != nil {
			b.handleConnectivityError(sid, err)
		} else {
			b.metricsMU.Lock()
			*op.ptr = c
			b.metricsMU.Unlock()
		}
	})
}

func (b *Agent) handleMetricsCollect(sid uint64, data []byte) {
	cmd := new(aranyagopb.MetricsCollectCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal MetricsCollectCmd: %w", err))
		return
	}

	type opTarget struct {
		name    string
		collect types.MetricsCollectFunc
	}

	b.metricsMU.RLock()

	op, ok := map[aranyagopb.MetricsTarget]*opTarget{
		aranyagopb.METRICS_TARGET_NODE:      {name: "node", collect: b.collectNodeMetrics},
		aranyagopb.METRICS_TARGET_CONTAINER: {name: "container", collect: b.collectContainerMetrics},
	}[cmd.Target]

	b.metricsMU.RUnlock()

	if !ok {
		b.handleUnknownCmd(sid, "metrics.collect", cmd)
		return
	}

	if op == nil {
		b.handleRuntimeError(sid, wellknownerrors.ErrNotSupported)
		return
	}

	b.processInNewGoroutine(sid, "metrics.collect."+op.name, func() {
		metricsData, err := op.collect()
		if err != nil {
			b.handleRuntimeError(sid, err)
			return
		}

		_, err = b.PostData(sid, aranyagopb.MSG_DATA_METRICS, 0, true, metricsData)
		if err != nil {
			b.handleConnectivityError(sid, err)
			return
		}
	})
}
