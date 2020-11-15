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
	"bytes"
	"fmt"
	"sync"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/wellknownerrors"
	dto "github.com/prometheus/client_model/go"

	"arhat.dev/arhat/pkg/metrics"
	"arhat.dev/arhat/pkg/metrics/metricsutils"
	"arhat.dev/arhat/pkg/types"
)

type agentComponentMetrics struct {
	metricsMU          *sync.RWMutex
	collectNodeMetrics types.MetricsCollectFunc
}

func (b *agentComponentMetrics) init() {
	b.metricsMU = new(sync.RWMutex)
}

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
			name:   "node",
			ptr:    &b.agentComponentMetrics.collectNodeMetrics,
			create: metrics.CreateNodeMetricsCollector,
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

func (b *Agent) encodeMetrics(metrics []*dto.MetricFamily) ([]byte, error) {
	buf := new(bytes.Buffer)
	gw := b.GetZstdWriter(buf)

	defer func() {
		_ = gw.Close()
		b.zstdPool.Put(gw)
	}()

	err := metricsutils.EncodeMetrics(gw, metrics)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
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
		aranyagopb.METRICS_TARGET_NODE: {name: "node", collect: b.collectNodeMetrics},
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
		mtc, err := op.collect()
		if err != nil {
			b.handleRuntimeError(sid, err)
			return
		}

		pMtcRaw := b.extensionComponentPeripheral.RetrieveCachedMetrics()
		pMtc, ok := pMtcRaw.([]*dto.MetricFamily)
		if ok {
			mtc = append(mtc, pMtc...)
		}

		data, err := b.encodeMetrics(mtc)
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
		metricsForNode, paramsForAgent, metricsForAgent := b.extensionComponentPeripheral.CollectMetrics(
			cmd.PeripheralNames...,
		)
		// TODO: add agent metrics report support
		_, _ = paramsForAgent, metricsForAgent

		b.extensionComponentPeripheral.CacheMetrics(metricsForNode)
	})
}
