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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateImageName(t *testing.T) {
	dockerHubImages := map[string]string{
		"alpine:latest":           "docker.io/library/alpine:latest",
		"library/alpine:latest":   "docker.io/library/alpine:latest",
		"user/foo:test":           "docker.io/user/foo:test",
		"docker.io/user/foo:test": "docker.io/user/foo:test",
	}
	for name, expectedName := range dockerHubImages {
		assert.Equal(t, expectedName, GenerateImageName(DefaultDockerImageDomain, DefaultDockerImageNamespace, name))
	}

	gcrImages := map[string]string{
		"k8s.gcr.io/pause:3.1": "k8s.gcr.io/pause:3.1",
		"pause:3.1":            "k8s.gcr.io/pause:3.1",
	}
	for name, expectedName := range gcrImages {
		assert.Equal(t, expectedName, GenerateImageName("k8s.gcr.io", "", name))
	}
}
