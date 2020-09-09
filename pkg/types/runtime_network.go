// +build !rt_none

package types

import "arhat.dev/aranya-proto/aranyagopb"

type NetworkRuntime interface {
	// UpdateContainerNetwork update cni config dynamically
	UpdateContainerNetwork(options *aranyagopb.NetworkOptions) ([]*aranyagopb.PodStatus, error)
}
