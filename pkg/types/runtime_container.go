// +build !rt_none

package types

import "arhat.dev/aranya-proto/aranyagopb"

type ContainerRuntime interface {
	// EnsureImages ensure container images
	EnsureImages(options *aranyagopb.ImageEnsureCmd) ([]*aranyagopb.ImageStatusMsg, error)

	// CreateContainers creates containers
	CreateContainers(options *aranyagopb.PodEnsureCmd) (*aranyagopb.PodStatusMsg, error)

	// DeleteContainer deletes a single container
	DeleteContainers(podUID string, containers []string) (*aranyagopb.PodStatusMsg, error)

	// DeletePod kills all containers and delete pod related volume data
	DeletePod(options *aranyagopb.PodDeleteCmd) (*aranyagopb.PodStatusMsg, error)

	// ListPods show (all) pods we are managing
	ListPods(options *aranyagopb.PodListCmd) ([]*aranyagopb.PodStatusMsg, error)
}
