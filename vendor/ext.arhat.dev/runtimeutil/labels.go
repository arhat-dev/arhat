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

package runtimeutil

import (
	"arhat.dev/aranya-proto/aranyagopb/runtimepb"
)

func AbbotMatchLabels() map[string]string {
	return map[string]string{
		ContainerLabelPodContainerRole: ContainerRoleWork,
		ContainerLabelPodContainer:     ContainerNameAbbot,
		LabelRole:                      LabelRoleValueAbbot,
		ContainerLabelHostNetwork:      "true",
	}
}

func IsPauseContainer(labels map[string]string) bool {
	if labels == nil {
		return false
	}

	return labels[ContainerLabelPodContainer] == ContainerNamePause
}

func IsAbbotPod(labels map[string]string) bool {
	if labels == nil {
		return false
	}

	// abbot container must use host network
	if !IsHostNetwork(labels) {
		return false
	}

	if labels[LabelRole] != LabelRoleValueAbbot {
		return false
	}

	return true
}

func IsHostNetwork(labels map[string]string) bool {
	if labels == nil {
		return false
	}

	_, ok := labels[ContainerLabelHostNetwork]
	return ok
}

func ContainerLabels(options *runtimepb.PodEnsureCmd, container string) map[string]string {
	defaults := map[string]string{
		ContainerLabelPodUID:       options.PodUid,
		ContainerLabelPodNamespace: options.Namespace,
		ContainerLabelPodName:      options.Name,
		ContainerLabelPodContainer: container,
		ContainerLabelPodContainerRole: func() string {
			switch container {
			case ContainerNamePause:
				return ContainerRoleInfra
			default:
				return ContainerRoleWork
			}
		}(),
	}

	result := make(map[string]string)
	for k, v := range options.Labels {
		result[k] = v
	}

	for k, v := range defaults {
		result[k] = v
	}

	if options.HostNetwork {
		result[ContainerLabelHostNetwork] = "true"
	}

	return result
}
