// +build nometrics

package metrics

import (
	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/wellknownerrors"

	"arhat.dev/arhat/pkg/types"
)

// CreateNodeMetricsCollector creates a new node metrics and a new container metrics collector
func CreateNodeMetricsCollector(config *aranyagopb.MetricsConfigCmd) (types.MetricsCollectFunc, error) {
	return nil, wellknownerrors.ErrNotSupported
}

func CreateContainerMetricsCollector(config *aranyagopb.MetricsConfigCmd) (types.MetricsCollectFunc, error) {
	return nil, wellknownerrors.ErrNotSupported
}
