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
	"fmt"

	"arhat.dev/aranya-proto/aranyagopb/runtimepb"
)

func GetContainerName(namespace, name, container string) string {
	return fmt.Sprintf("%s.%s.%s", namespace, name, container)
}

// nolint:goconst
func SharedNamespaces(pauseCtrID string, options *runtimepb.PodEnsureCmd) map[string]string {
	containerNS := fmt.Sprintf("container:%s", pauseCtrID)
	ns := map[string]string{
		"net":  containerNS,
		"user": containerNS,
		"ipc":  containerNS,
		"uts":  containerNS,
	}

	if options.HostNetwork {
		ns["net"] = "host"
	}

	if options.HostIpc {
		ns["ipc"] = "host"
	}

	if options.HostPid {
		ns["pid"] = "host"
	} else if options.SharePid {
		ns["pid"] = containerNS
	}

	return ns
}
