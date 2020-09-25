package connectivity

import (
	"arhat.dev/aranya-proto/aranyagopb"

	"arhat.dev/arhat/pkg/types"
)

// nolint:lll
type FactoryFunc func(
	target string,
	params map[string]string,
	tlsConfig *aranyagopb.TLSConfig,
) (types.Connectivity, error)

type key struct {
	name string
	mode aranyagopb.ConnectivityMode
}

var (
	supportedConnectivity = make(map[key]FactoryFunc)
)

// Register add one connectivity method for device
func Register(
	name string,
	mode aranyagopb.ConnectivityMode,
	factory FactoryFunc,
) {
	supportedConnectivity[key{
		name: name,
		mode: mode,
	}] = factory
}

func NewConnectivity(
	name string,
	mode aranyagopb.ConnectivityMode,
) (FactoryFunc, bool) {
	factory, ok := supportedConnectivity[key{
		name: name,
		mode: mode,
	}]

	return factory, ok
}
