// +build !nometrics
// +build !windows

package metrics

import (
	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/wellknownerrors"

	"arhat.dev/arhat/pkg/metrics/unixexporter"
	"arhat.dev/arhat/pkg/types"
)

// CreateNodeMetricsCollector creates a new node metrics and a new container metrics collector
func CreateNodeMetricsCollector(config *aranyagopb.MetricsConfigOptions) (types.MetricsCollectFunc, error) {
	if len(config.Collect) == 0 {
		return nil, wellknownerrors.ErrInvalidOperation
	}

	return unixexporter.CreateNodeMetricsGatherer(config)
}

func CreateContainerMetricsCollector(config *aranyagopb.MetricsConfigOptions) (types.MetricsCollectFunc, error) {
	if len(config.Collect) == 0 {
		return nil, wellknownerrors.ErrInvalidOperation
	}

	return unixexporter.CreateContainerMetricsGatherer(config)
}
