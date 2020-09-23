// +build !rt_none

package types

import "arhat.dev/aranya-proto/aranyagopb"

type RuntimeContainer interface {
	// EnsurePod creates containers
	EnsurePod(options *aranyagopb.PodEnsureCmd) (*aranyagopb.PodStatusMsg, error)

	// DeletePod kills all containers and delete pod related volume data
	DeletePod(options *aranyagopb.PodDeleteCmd) (*aranyagopb.PodStatusMsg, error)

	// ListPods show (all) pods we are managing
	ListPods(options *aranyagopb.PodListCmd) ([]*aranyagopb.PodStatusMsg, error)
}

type RuntimeImage interface {
	// EnsureImages ensure container images
	EnsureImages(options *aranyagopb.ImageEnsureCmd) ([]*aranyagopb.ImageStatusMsg, error)
}

type RuntimeContainerNetwork interface {
	// UpdateContainerNetwork update cni config dynamically
	UpdateContainerNetwork(options *aranyagopb.ContainerNetworkEnsureCmd) ([]*aranyagopb.PodStatusMsg, error)
}
