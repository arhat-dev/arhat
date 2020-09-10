// +build !nodevice

package agent

// nolint:golint
import (
	_ "arhat.dev/arhat/pkg/devices/coap"
	_ "arhat.dev/arhat/pkg/devices/file"
	_ "arhat.dev/arhat/pkg/devices/gpio"
	_ "arhat.dev/arhat/pkg/devices/http"
	_ "arhat.dev/arhat/pkg/devices/modbus"
	_ "arhat.dev/arhat/pkg/devices/mqtt"
	_ "arhat.dev/arhat/pkg/devices/opcua"
	_ "arhat.dev/arhat/pkg/devices/serial"
)
