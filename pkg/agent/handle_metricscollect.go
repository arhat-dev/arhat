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

	"arhat.dev/arhat/pkg/types"
)

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
		metrics, err := op.collect()
		if err != nil {
			b.handleRuntimeError(sid, err)
			return
		}

		metrics = append(metrics, b.peripheralManager.RetrieveCachedMetrics()...)
		data, err := b.encodeMetrics(metrics)
		if err != nil {
			b.handleConnectivityError(sid, fmt.Errorf("failed to encode metrics: %w", err))
			return
		}

		_, err = b.PostData(sid, aranyagopb.MSG_DATA_METRICS, 0, true, data)
		if err != nil {
			b.handleConnectivityError(sid, err)
			return
		}
	})
}

func (b *Agent) handlePeripheralMetricsCollect(sid uint64, data []byte) {
	cmd := new(aranyagopb.PeripheralMetricsCollectCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal PeripheralMetricsCollectCmd: %w", err))
		return
	}

	b.processInNewGoroutine(sid, "peripheral.metrics", func() {
		metricsForNode, paramsForAgent, metricsForAgent := b.peripheralManager.CollectMetrics(cmd.PeripheralNames...)
		_, _ = paramsForAgent, metricsForAgent
		// TODO: add agent metrics report support

		b.peripheralManager.CacheMetrics(metricsForNode)
	})
}
