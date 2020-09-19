// +build !rt_none

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
	"arhat.dev/aranya-proto/aranyagopb"

	"arhat.dev/arhat/pkg/constant"
)

func AbbotMatchLabels() map[string]string {
	return map[string]string{
		constant.ContainerLabelPodContainerRole: constant.ContainerRoleWork,
		constant.ContainerLabelPodContainer:     constant.ContainerNameAbbot,
		constant.LabelRole:                      constant.LabelRoleValueAbbot,
		constant.ContainerLabelHostNetwork:      "true",
	}
}

func IsPauseContainer(labels map[string]string) bool {
	if labels == nil {
		return false
	}

	return labels[constant.ContainerLabelPodContainer] == constant.ContainerNamePause
}

func IsAbbotPod(labels map[string]string) bool {
	if labels == nil {
		return false
	}

	// abbot container must use host network
	if !IsHostNetwork(labels) {
		return false
	}

	if labels[constant.LabelRole] != constant.LabelRoleValueAbbot {
		return false
	}

	return true
}

func IsHostNetwork(labels map[string]string) bool {
	if labels == nil {
		return false
	}

	_, ok := labels[constant.ContainerLabelHostNetwork]
	return ok
}

func ContainerLabels(options *aranyagopb.PodEnsureCmd, container string) map[string]string {
	defaults := map[string]string{
		constant.ContainerLabelPodUID:       options.PodUid,
		constant.ContainerLabelPodNamespace: options.Namespace,
		constant.ContainerLabelPodName:      options.Name,
		constant.ContainerLabelPodContainer: container,
		constant.ContainerLabelPodContainerRole: func() string {
			switch container {
			case constant.ContainerNamePause:
				return constant.ContainerRoleInfra
			default:
				return constant.ContainerRoleWork
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
		result[constant.ContainerLabelHostNetwork] = "true"
	}

	return result
}
