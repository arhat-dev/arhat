// +build !nometrics

/*
Copyright 2019 The arhat.dev Authors.

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

func (b *Agent) handleMetricsCmd(sid uint64, data []byte) {
	cmd := new(aranyagopb.MetricsCmd)

	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal metrics cmd: %w", err))
		return
	}

	switch cmd.Action {
	case aranyagopb.CONFIGURE_NODE_METRICS_COLLECTION:
		config := cmd.GetConfig()
		if config == nil {
			b.handleRuntimeError(sid, errRequiredOptionsNotFound)
		}

		b.processInNewGoroutine(sid, "metrics.config.node", func() {
			b.doConfigureNodeMetrics(sid, config)
		})
	case aranyagopb.CONFIGURE_CONTAINER_METRICS_COLLECTION:
		config := cmd.GetConfig()
		if config == nil {
			b.handleRuntimeError(sid, errRequiredOptionsNotFound)
		}

		b.processInNewGoroutine(sid, "metrics.config.container", func() {
			b.doConfigureContainerMetrics(sid, config)
		})
	case aranyagopb.COLLECT_NODE_METRICS:
		b.processInNewGoroutine(sid, "metrics.collect.node", func() {
			b.doCollectNodeMetrics(sid)
		})
	case aranyagopb.COLLECT_CONTAINER_METRICS:
		b.processInNewGoroutine(sid, "metrics.collect.container", func() {
			b.doCollectContainerMetrics(sid)
		})
	default:
		b.handleUnknownCmd(sid, "metrics", cmd)
	}
}

func (b *Agent) doConfigureNodeMetrics(sid uint64, config *aranyagopb.MetricsConfigOptions) {
	c, err := metrics.CreateNodeMetricsCollector(config)
	if err != nil {
		b.handleRuntimeError(sid, err)
		return
	}

	err = b.PostMsg(aranyagopb.NewMetricsConfigured(sid))
	if err != nil {
		b.handleConnectivityError(sid, err)
	} else {
		b.metricsMU.Lock()
		b.collectNodeMetrics = c
		b.metricsMU.Unlock()
	}
}

func (b *Agent) doConfigureContainerMetrics(sid uint64, config *aranyagopb.MetricsConfigOptions) {
	c, err := metrics.CreateContainerMetricsCollector(config)
	if err != nil {
		b.handleRuntimeError(sid, err)
		return
	}

	err = b.PostMsg(aranyagopb.NewMetricsConfigured(sid))
	if err != nil {
		b.handleConnectivityError(sid, err)
	} else {
		b.metricsMU.Lock()
		b.collectContainerMetrics = c
		b.metricsMU.Unlock()
	}
}

func (b *Agent) doCollectNodeMetrics(sid uint64) {
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

	if err := b.PostMsg(aranyagopb.NewNodeMetrics(sid, metricsData)); err != nil {
		b.handleConnectivityError(sid, err)
	}
}

func (b *Agent) doCollectContainerMetrics(sid uint64) {
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

	if err := b.PostMsg(aranyagopb.NewContainerMetrics(sid, metricsData)); err != nil {
		b.handleConnectivityError(sid, err)
	}
}
