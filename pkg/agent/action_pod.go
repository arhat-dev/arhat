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

func (b *Agent) handlePodList(sid uint64, data []byte) {
	cmd := new(aranyagopb.PodListCmd)

	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal PodListCmd: %w", err))
		return
	}

	rc, ok := b.runtime.(types.RuntimeContainer)
	if !ok {
		b.handleUnknownCmd(sid, "pod.list", cmd)
		return
	}

	b.processInNewGoroutine(sid, "pod.list", func() {
		pods, err := rc.ListPods(cmd)
		if err != nil {
			b.handleRuntimeError(sid, err)
			return
		}

		err = b.PostMsg(sid, aranyagopb.MSG_POD_STATUS_LIST, aranyagopb.NewPodStatusListMsg(pods))
		if err != nil {
			b.handleConnectivityError(sid, err)
			return
		}
	})
}

func (b *Agent) handlePodEnsure(sid uint64, data []byte) {
	cmd := new(aranyagopb.PodEnsureCmd)

	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal PodEnsureCmd: %w", err))
		return
	}

	rc, ok := b.runtime.(types.RuntimeContainer)
	if !ok {
		b.handleUnknownCmd(sid, "pod.ensure", cmd)
		return
	}

	b.processInNewGoroutine(sid, "pod.ensure", func() {
		podStatus, err := rc.EnsurePod(cmd)
		if err != nil {
			b.handleRuntimeError(sid, err)
			return
		}

		if err := b.PostMsg(sid, aranyagopb.MSG_POD_STATUS, podStatus); err != nil {
			b.handleConnectivityError(sid, err)
			return
		}
	})
}

func (b *Agent) handlePodDelete(sid uint64, data []byte) {
	cmd := new(aranyagopb.PodDeleteCmd)

	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal PodDeleteCmd: %w", err))
		return
	}

	rc, ok := b.runtime.(types.RuntimeContainer)
	if !ok {
		b.handleUnknownCmd(sid, "pod.delete", cmd)
		return
	}

	b.processInNewGoroutine(sid, "pod.delete", func() {
		podDeleted, err := rc.DeletePod(cmd)
		if err != nil {
			b.handleRuntimeError(sid, err)
			return
		}

		if err := b.PostMsg(sid, aranyagopb.MSG_POD_STATUS, podDeleted); err != nil {
			b.handleConnectivityError(sid, err)
			return
		}
	})
}
