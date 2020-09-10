/*
Copyright 2020 The arhat.dev Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package constant

const (
	LabelRole = "arhat.dev/role"

	LabelRoleValueAbbot = "Abbot"
)

// labels used by container runtime
const (
	ContainerLabelPodNamespace     = "container.arhat.dev/pod-namespace"
	ContainerLabelPodName          = "container.arhat.dev/pod-name"
	ContainerLabelPodUID           = "container.arhat.dev/pod-uid"
	ContainerLabelHostNetwork      = "config.container.arhat.dev/host-network"
	ContainerLabelPodContainer     = "container.arhat.dev/pod-container"
	ContainerLabelPodContainerRole = "container.arhat.dev/pod-container-role"

	ContainerRoleInfra = "infra"
	ContainerRoleWork  = "work"
)
