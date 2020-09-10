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
	"path/filepath"
	"strings"
)

const (
	DefaultDockerImageDomain    = "docker.io"
	DefaultDockerImageNamespace = "library"
)

// GenerateImageName create a image name with defaults according to provided name
// defaultDomain MUST NOT be empty
func GenerateImageName(defaultDomain, defaultNamespace, name string) string {
	firstSlashIndex := strings.IndexByte(name, '/')
	switch firstSlashIndex {
	case -1:
		// no slash, add default registry
		return filepath.Join(defaultDomain, defaultNamespace, name)
	default:
		prefix := name[:firstSlashIndex]
		if strings.Contains(prefix, ".") {
			return name
		}
		return filepath.Join(defaultDomain, name)
	}
}

func GetEnv(env []string) map[string]string {
	result := make(map[string]string)
	for _, kv := range env {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		} else {
			result[parts[0]] = ""
		}
	}

	return result
}
