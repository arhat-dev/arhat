// +build !nodev

package types

import "arhat.dev/aranya-proto/aranyagopb"

// nolint:lll
type NewDeviceConnectivityFunc func(target string, params map[string]string, tlsConfig *aranyagopb.DeviceConnectivityTLSConfig) (DeviceConnectivity, error)

type DeviceConnectivity interface {
	Connect() error
	Operate(params map[string]string) ([][]byte, error)
	CollectMetrics(params map[string]string) ([][]byte, error)
	Close() error
}

var (
	clientModeConnectivity = make(map[string]NewDeviceConnectivityFunc)
	serverModeConnectivity = make(map[string]NewDeviceConnectivityFunc)
)

// RegisterDeviceTransport add one connectivity method to supported method list
func RegisterDeviceConnectivityFactory(
	name string,
	mode aranyagopb.DeviceConnectivity_Mode,
	factory NewDeviceConnectivityFunc,
) {
	switch mode {
	case aranyagopb.DEVICE_CONNECTIVITY_MODE_CLIENT:
		clientModeConnectivity[name] = factory
	case aranyagopb.DEVICE_CONNECTIVITY_MODE_SERVER:
		serverModeConnectivity[name] = factory
	}
}

func GetDeviceConnectivityFactory(
	name string,
	mode aranyagopb.DeviceConnectivity_Mode,
) (NewDeviceConnectivityFunc, bool) {
	var (
		factory NewDeviceConnectivityFunc
		ok      bool
	)

	switch mode {
	case aranyagopb.DEVICE_CONNECTIVITY_MODE_CLIENT:
		factory, ok = clientModeConnectivity[name]
	case aranyagopb.DEVICE_CONNECTIVITY_MODE_SERVER:
		factory, ok = serverModeConnectivity[name]
	}

	return factory, ok
}
