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
	"github.com/klauspost/compress/zstd"
	dto "github.com/prometheus/client_model/go"

	"arhat.dev/arhat/pkg/metrics"
	"arhat.dev/arhat/pkg/metrics/metricsutils"
)

type agentComponentMetrics struct {
	zstdEncoder *zstd.Encoder

	metricsMU          *sync.RWMutex
	collectNodeMetrics metrics.CollectFunc
}

func (b *agentComponentMetrics) init() error {
	var err error
	b.zstdEncoder, err = zstd.NewWriter(
		nil,
		zstd.WithEncoderLevel(zstd.SpeedBestCompression),
	)
	if err != nil {
		return fmt.Errorf("failed to create zstd encoder: %w", err)
	}

	b.metricsMU = new(sync.RWMutex)

	return nil
}

func (b *Agent) handleMetricsConfig(sid uint64, _ *uint32, data []byte) {
	cmd := new(aranyagopb.MetricsConfigCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal MetricsConfigCmd: %w", err))
		return
	}

	b.metricsMU.Lock()
	defer b.metricsMU.Unlock()

	c, err := metrics.CreateCollector(cmd)
	if err != nil {
		b.handleRuntimeError(sid, err)
		return
	}

	err = b.PostMsg(sid, aranyagopb.MSG_DONE, nil)
	if err != nil {
		b.handleConnectivityError(sid, err)
	} else {
		b.agentComponentMetrics.collectNodeMetrics = c
	}
}

func (b *Agent) encodeMetrics(metrics []*dto.MetricFamily) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := metricsutils.EncodeMetrics(buf, metrics)
	if err != nil {
		return nil, err
	}

	return b.zstdEncoder.EncodeAll(buf.Bytes(), nil), nil
}

func (b *Agent) handleMetricsCollect(sid uint64, _ *uint32, data []byte) {
	// data is ignore and shold be nil
	_ = data

	b.metricsMU.RLock()
	collect := b.collectNodeMetrics
	b.metricsMU.RUnlock()

	if collect == nil {
		b.handleRuntimeError(sid, fmt.Errorf("node metrics collector not configured"))
		return
	}

	b.processInNewGoroutine(sid, "metrics.collect", func() {
		mtc, err := collect()
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
