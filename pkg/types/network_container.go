// +build !rt_none

package types

import "arhat.dev/aranya-proto/aranyagopb"

type ContainerNetworkClient interface {
	NetworkClient

	// EnsureContainerNetwork to update cni config with provided options
	EnsureContainerNetwork(options *aranyagopb.ContainerNetworkEnsureCmd) error

	// EnsurePodNetwork to create/update single pod network
	EnsurePodNetwork(
		namespace, name string, ctrID string, pid uint32, opts *aranyagopb.PodNetworkSpec,
	) (ipv4, ipv6 string, err error)

	// RestorePodNetwork to restore single pod network from file
	RestorePodNetwork(ctrID string, pid uint32) error

	// DeletePodNetwork to delete single pod network
	DeletePodNetwork(ctrID string, pid uint32) error

	// GetPodIPAddresses to retrieve pod ip addresses
	GetPodIPAddresses(pid uint32) (ipv4, ipv6 string, err error)
}
