// +build !nometrics
// +build !noextension,!noextension_peripheral

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
)

func (b *Agent) handlePeripheralMetricsCollect(sid uint64, data []byte) {
	if b.Manager == nil {
		b.handleUnknownCmd(sid, "peripheral.metrics", nil)
		return
	}

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
