// +build !nometrics

package metrics

import (
	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/wellknownerrors"

	"arhat.dev/arhat/pkg/metrics/windowsexporter"
	"arhat.dev/arhat/pkg/types"
)

// CreateNodeMetricsCollector creates a new node metrics and a new container metrics collector
func CreateNodeMetricsCollector(config *aranyagopb.MetricsConfigCmd) (types.MetricsCollectFunc, error) {
	if len(config.Collect) == 0 {
		return nil, wellknownerrors.ErrInvalidOperation
	}

	return windowsexporter.CreateNodeMetricsGatherer(config)
}

func CreateContainerMetricsCollector(config *aranyagopb.MetricsConfigCmd) (types.MetricsCollectFunc, error) {
	if len(config.Collect) == 0 {
		return nil, wellknownerrors.ErrInvalidOperation
	}

	return windowsexporter.CreateContainerMetricsGatherer(config)
}
