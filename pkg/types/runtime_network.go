// +build !rt_none

package types

import "arhat.dev/aranya-proto/aranyagopb"

type NetworkRuntime interface {
	// UpdateContainerNetwork update cni config dynamically
	UpdateContainerNetwork(options *aranyagopb.NetworkUpdatePodNetworkCmd) ([]*aranyagopb.PodStatusMsg, error)
}
