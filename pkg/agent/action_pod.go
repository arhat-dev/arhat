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

package agent

import (
	"fmt"

	"arhat.dev/aranya-proto/aranyagopb"

	"arhat.dev/arhat/pkg/types"
)

func (b *Agent) handlePodCmd(sid uint64, data []byte) {
	cmd := new(aranyagopb.PodCmd)

	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal pod cmd: %w", err))
		return
	}

	switch cmd.Action {
	case aranyagopb.START_POD_SYNC_LOOP:
		b.handleSyncLoop(sid, "pod.syncLoop", cmd.GetSyncOptions(), func() {
			b.doPodList(0, &aranyagopb.ListOptions{All: true})
		})
	case aranyagopb.ENSURE_IMAGES:
		imageEnsureOpt := cmd.GetImageEnsureOptions()
		if imageEnsureOpt == nil {
			b.handleRuntimeError(sid, errRequiredOptionsNotFound)
			return
		}

		b.processInNewGoroutine(sid, "pod.ensureImage", func() {
			b.doPodEnsureImages(sid, imageEnsureOpt)
		})
	case aranyagopb.CREATE_CONTAINERS:
		createOpt := cmd.GetCreateOptions()
		if createOpt == nil {
			b.handleRuntimeError(sid, errRequiredOptionsNotFound)
			return
		}

		b.processInNewGoroutine(sid, "pod.createContainers", func() {
			b.doContainerCreate(sid, createOpt)
		})
	case aranyagopb.DELETE_POD:
		deleteOpt := cmd.GetDeleteOptions()
		if deleteOpt == nil {
			b.handleRuntimeError(sid, errRequiredOptionsNotFound)
			return
		}

		b.processInNewGoroutine(sid, "pod.delete", func() {
			b.doPodDelete(sid, deleteOpt)
		})
	case aranyagopb.LIST_PODS:
		listOpt := cmd.GetListOptions()
		if listOpt == nil {
			b.handleRuntimeError(sid, errRequiredOptionsNotFound)
			return
		}

		b.processInNewGoroutine(sid, "pod.list", func() {
			b.doPodList(sid, listOpt)
		})
	case aranyagopb.DELETE_CONTAINERS:
		deleteOpt := cmd.GetDeleteOptions()
		if deleteOpt == nil {
			b.handleRuntimeError(sid, errRequiredOptionsNotFound)
			return
		}

		b.processInNewGoroutine(sid, "ctr.delete", func() {
			b.doContainerDelete(sid, deleteOpt)
		})
	default:
		b.handleUnknownCmd(sid, "pod", cmd)
		return
	}
}

func (b *Agent) doPodEnsureImages(sid uint64, options *aranyagopb.ImageEnsureOptions) {
	pulledImages, err := b.runtime.(types.ContainerRuntime).EnsureImages(options)
	if err != nil {
		b.handleRuntimeError(sid, err)
		return
	}

	if err := b.PostMsg(aranyagopb.NewImageListMsg(sid, pulledImages)); err != nil {
		b.handleConnectivityError(sid, err)
		return
	}
}

func (b *Agent) doContainerCreate(sid uint64, options *aranyagopb.CreateOptions) {
	podStatus, err := b.runtime.(types.ContainerRuntime).CreateContainers(options)
	if err != nil {
		b.handleRuntimeError(sid, err)
		return
	}

	if err := b.PostMsg(aranyagopb.NewPodStatusMsg(sid, podStatus)); err != nil {
		b.handleConnectivityError(sid, err)
		return
	}
}

func (b *Agent) doContainerDelete(sid uint64, options *aranyagopb.DeleteOptions) {
	s, err := b.runtime.(types.ContainerRuntime).DeleteContainers(options.PodUid, options.Containers)
	if err != nil {
		b.handleRuntimeError(sid, err)
		return
	}

	if err := b.PostMsg(aranyagopb.NewPodStatusMsg(sid, s)); err != nil {
		b.handleConnectivityError(sid, err)
	}
}

func (b *Agent) doPodDelete(sid uint64, options *aranyagopb.DeleteOptions) {
	podDeleted, err := b.runtime.(types.ContainerRuntime).DeletePod(options)
	if err != nil {
		b.handleRuntimeError(sid, err)
		return
	}

	if err := b.PostMsg(aranyagopb.NewPodStatusMsg(sid, podDeleted)); err != nil {
		b.handleConnectivityError(sid, err)
		return
	}
}

func (b *Agent) doPodList(sid uint64, options *aranyagopb.ListOptions) {
	pods, err := b.runtime.(types.ContainerRuntime).ListPods(options)
	if err != nil {
		b.handleRuntimeError(sid, err)
		return
	}

	if err := b.PostMsg(aranyagopb.NewPodStatusListMsg(sid, pods)); err != nil {
		b.handleConnectivityError(sid, err)
		return
	}
}
