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
	"context"
	"path/filepath"
	"time"
)

func NewBaseRuntime(
	ctx context.Context,
	dataDir string,
	imageActionTimeout, podActionTimeout time.Duration,
	name, version, os, osImage, arch, kernelVersion string,
) *BaseRuntime {
	return &BaseRuntime{
		dataDir: dataDir,

		imageActionTimeout: imageActionTimeout,
		podActionTimeout:   podActionTimeout,

		name:          name,
		version:       version,
		os:            os,
		osImage:       osImage,
		arch:          arch,
		kernelVersion: kernelVersion,
	}
}

type BaseRuntime struct {
	dataDir string

	podActionTimeout   time.Duration
	imageActionTimeout time.Duration

	name, version, os, osImage,
	arch, kernelVersion string
}

func (r *BaseRuntime) Name() string          { return r.name }
func (r *BaseRuntime) Version() string       { return r.version }
func (r *BaseRuntime) OS() string            { return r.os }
func (r *BaseRuntime) OSImage() string       { return r.osImage }
func (r *BaseRuntime) Arch() string          { return r.arch }
func (r *BaseRuntime) KernelVersion() string { return r.kernelVersion }

func (r *BaseRuntime) ImageActionContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, r.imageActionTimeout)
}

func (r *BaseRuntime) PodActionContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, r.podActionTimeout)
}

func (r *BaseRuntime) ActionContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithCancel(ctx)
}

func (r *BaseRuntime) PodDir(podUID string) string {
	return filepath.Join(r.dataDir, "pods", podUID)
}

func (r *BaseRuntime) podVolumeDir(podUID, typ, volumeName string) string {
	return filepath.Join(r.PodDir(podUID), "volumes", typ, volumeName)
}

func (r *BaseRuntime) PodRemoteVolumeDir(podUID, volumeName string) string {
	return r.podVolumeDir(podUID, "remote", volumeName)
}

func (r *BaseRuntime) PodBindVolumeDir(podUID, volumeName string) string {
	return r.podVolumeDir(podUID, "bind", volumeName)
}

func (r *BaseRuntime) PodTmpfsVolumeDir(podUID, volumeName string) string {
	return r.podVolumeDir(podUID, "tmpfs", volumeName)
}

func (r *BaseRuntime) PodResolvConfFile(podUID string) string {
	return filepath.Join(r.PodDir(podUID), "volumes", "bind", "_net", "resolv.conf")
}
