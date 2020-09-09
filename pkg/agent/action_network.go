// +build !rt_none

/*
Copyright 2019 The arhat.dev Authors.

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

func (b *Agent) handleNetworkCmd(sid uint64, data []byte) {
	cmd := new(aranyagopb.NetworkCmd)

	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal network cmd: %w", err))
		return
	}

	switch cmd.Action {
	case aranyagopb.UPDATE_NETWORK:
		b.processInNewGoroutine(sid, "net.update", func() {
			b.doNetworkUpdate(sid, cmd.GetNetworkOptions())
		})
	default:
		b.handleUnknownCmd(sid, "net", cmd)
	}
}

func (b *Agent) doNetworkUpdate(sid uint64, options *aranyagopb.NetworkOptions) {
	result, err := b.runtime.(types.NetworkRuntime).UpdateContainerNetwork(options)
	if err != nil {
		b.handleRuntimeError(sid, err)
	}

	if err := b.PostMsg(aranyagopb.NewPodStatusListMsg(sid, result)); err != nil {
		b.handleConnectivityError(sid, err)
	}
}
