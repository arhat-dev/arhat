// +build !nometrics
// +build !windows

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

package metrics

import (
	"fmt"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/wellknownerrors"

	"arhat.dev/arhat/pkg/metrics/unixexporter"
	"arhat.dev/arhat/pkg/types"
)

// CreateNodeMetricsCollector creates a new node metrics and a new container metrics collector
func CreateNodeMetricsCollector(config *aranyagopb.MetricsConfigCmd) (types.MetricsCollectFunc, error) {
	if len(config.Collect) == 0 {
		return nil, wellknownerrors.ErrInvalidOperation
	}

	g, err := unixexporter.CreateNodeMetricsGatherer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create node metrics gatherer: %w", err)
	}

	return g.Gather, nil
}
