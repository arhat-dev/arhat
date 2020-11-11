// +build !nometrics
// +build !windows

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
