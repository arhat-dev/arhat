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

package libpod

import (
	"github.com/containers/libpod/v2/libpod"
)

func filterLabels(filters, labels map[string]string) bool {
	if len(labels) >= len(filters) {
		match := true
		for k, v := range filters {
			if labels[k] != v {
				match = false
				break
			}
		}
		return match
	}
	return false
}

func containerLabelFilterFunc(filters map[string]string) func(*libpod.Container) bool {
	return func(ctr *libpod.Container) bool {
		return filterLabels(filters, ctr.Labels())
	}
}

func podLabelFilterFunc(filters map[string]string) func(*libpod.Pod) bool {
	return func(pod *libpod.Pod) bool {
		return filterLabels(filters, pod.Labels())
	}
}
