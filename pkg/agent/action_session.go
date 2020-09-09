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
	"arhat.dev/pkg/log"
)

func (b *Agent) handleSessionCmd(sid uint64, data []byte) {
	cmd := new(aranyagopb.SessionCmd)

	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal session cmd: %w", err))
		return
	}

	switch cmd.Action {
	case aranyagopb.CLOSE_SESSION:
		b.logger.D("close session", log.Uint64("sid", sid))
		b.streams.Close(sid)
	default:
		b.handleUnknownCmd(sid, "close", cmd)
	}
}
