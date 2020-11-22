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
)

func (b *Agent) handleNetwork(sid uint64, _ *uint32, data []byte) {
	cmd := new(aranyagopb.NetworkCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal NetworkCmd: %w", err))
		return
	}

	b.processInNewGoroutine(sid, "net", func() {
		respBytes, err2 := b.networkClient.Do(b.ctx, cmd.AbbotRequestBytes, 0, "")
		if err2 != nil {
			b.handleConnectivityError(sid, err2)
			return
		}

		err2 = b.PostMsg(sid, aranyagopb.MSG_NET, &aranyagopb.NetworkMsg{
			AbbotResponseBytes: respBytes,
		})
		if err2 != nil {
			b.handleConnectivityError(sid, err2)
		}
	})
}
