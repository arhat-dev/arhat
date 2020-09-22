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
	"arhat.dev/pkg/log"
)

func (b *Agent) handleRejectCmd(sid uint64, data []byte) {
	cmd := new(aranyagopb.RejectCmd)

	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal RejectCmd: %w", err))
		return
	}

	switch cmd.Reason {
	case aranyagopb.REJECTION_ALREADY_CONNECTED,
		aranyagopb.REJECTION_POD_STATUS_SYNC_ERROR,
		aranyagopb.REJECTION_NODE_STATUS_SYNC_ERROR,
		aranyagopb.REJECTION_NETWORK_UPDATE_FAILURE,
		aranyagopb.REJECTION_CREDENTIAL_FAILURE,
		aranyagopb.REJECTION_INTERNAL_SERVER_ERROR:
		b.logger.D("rejected by server", log.String("reason", cmd.Reason.String()), log.String("msg", cmd.Message))
	default:
		b.handleUnknownCmd(sid, "reject", cmd)
	}

	if c := b.GetClient(); c != nil {
		_ = c.Close()
	}
}
