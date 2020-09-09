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
	"arhat.dev/pkg/hashhelper"
)

func (b *Agent) handleCredentialCmd(sid uint64, data []byte) {
	cmd := new(aranyagopb.CredentialCmd)

	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal credential cmd: %w", err))
		return
	}

	switch cmd.Action {
	case aranyagopb.UPDATE_STORAGE_CREDENTIAL:
		storageCred := cmd.GetStorage()
		if storageCred == nil {
			b.handleRuntimeError(sid, errRequiredOptionsNotFound)
			return
		}

		b.processInNewGoroutine(sid, "storage.credential.update", func() {
			b.doStorageCredentialUpdate(sid, storageCred)
		})
	default:
		b.handleUnknownCmd(sid, "credential", cmd)
	}
}

func (b *Agent) doStorageCredentialUpdate(sid uint64, cred *aranyagopb.StorageCredentialOptions) {
	msg := aranyagopb.NewCredentialStatus(sid, hashhelper.Sha256SumHex(cred.SshPrivateKey))
	if err := b.PostMsg(msg); err != nil {
		b.handleConnectivityError(sid, err)
		// do not make early return, let server make reject decision
	}

	b.storage.SetCredentials(cred)
}
