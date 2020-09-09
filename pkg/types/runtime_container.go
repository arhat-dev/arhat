// +build !rt_none

package types

import "arhat.dev/aranya-proto/aranyagopb"

type ContainerRuntime interface {
	// EnsureImages ensure container images
	EnsureImages(options *aranyagopb.ImageEnsureOptions) ([]*aranyagopb.Image, error)

	// CreateContainers creates containers
	CreateContainers(options *aranyagopb.CreateOptions) (*aranyagopb.PodStatus, error)

	// DeleteContainer deletes a single container
	DeleteContainers(podUID string, containers []string) (*aranyagopb.PodStatus, error)

	// DeletePod kills all containers and delete pod related volume data
	DeletePod(options *aranyagopb.DeleteOptions) (*aranyagopb.PodStatus, error)

	// ListPods show (all) pods we are managing
	ListPods(options *aranyagopb.ListOptions) ([]*aranyagopb.PodStatus, error)
}
