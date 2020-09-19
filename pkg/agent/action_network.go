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

func (b *Agent) handleNetworkUpdatePodNet(sid uint64, data []byte) {
	cmd := new(aranyagopb.NetworkUpdatePodNetworkCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal NetworkUpdatePodNetworkCmd: %w", err))
		return
	}

	b.processInNewGoroutine(sid, "net.update", func() {
		result, err := b.runtime.(types.NetworkRuntime).UpdateContainerNetwork(cmd)
		if err != nil {
			b.handleRuntimeError(sid, err)
		}

		msg := aranyagopb.NewPodStatusListMsg(result)
		err = b.PostMsg(sid, aranyagopb.MSG_NETWORK_STATUS, msg)
		if err != nil {
			b.handleConnectivityError(sid, err)
		}
	})
}
