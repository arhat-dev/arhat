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
	"sync"

	"arhat.dev/pkg/log"

	"arhat.dev/arhat/pkg/conf"
)

func NewBaseRuntime(
	ctx context.Context,
	config *conf.ArhatRuntimeConfig,
	name, version, os, arch, kernelVersion string,
) *BaseRuntime {
	return &BaseRuntime{
		ArhatRuntimeConfig: config,

		ctx:           ctx,
		name:          name,
		version:       version,
		os:            os,
		arch:          arch,
		kernelVersion: kernelVersion,

		logger: log.Log.WithName("runtime"),

		mu: new(sync.RWMutex),
	}
}

type BaseRuntime struct {
	*conf.ArhatRuntimeConfig

	ctx    context.Context
	logger log.Interface

	name, version, os, arch,
	kernelVersion string

	mu *sync.RWMutex
}

func (b *BaseRuntime) Log() log.Interface {
	return b.logger
}

func (b *BaseRuntime) OS() string {
	return b.os
}

func (b *BaseRuntime) Arch() string {
	return b.arch
}

func (b *BaseRuntime) KernelVersion() string {
	return b.kernelVersion
}

func (b *BaseRuntime) Name() string {
	return b.name
}

func (b *BaseRuntime) Version() string {
	return b.version
}

func (b *BaseRuntime) SetContext(ctx context.Context) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.ctx = ctx
}

func (b *BaseRuntime) ImageActionContext() (context.Context, context.CancelFunc) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return context.WithTimeout(b.ctx, b.ArhatRuntimeConfig.EndPoints.Image.ActionTimeout)
}

func (b *BaseRuntime) RuntimeActionContext() (context.Context, context.CancelFunc) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return context.WithTimeout(b.ctx, b.ArhatRuntimeConfig.EndPoints.Container.ActionTimeout)
}

func (b *BaseRuntime) ActionContext() (context.Context, context.CancelFunc) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return context.WithCancel(b.ctx)
}
