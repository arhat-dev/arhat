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

package types

import (
	"arhat.dev/aranya-proto/aranyagopb"
	"github.com/gogo/protobuf/proto"
)

type AgentCmdHandleFunc func(cmd *aranyagopb.Cmd)

type Agent interface {
	// HandleCmd received from aranya
	HandleCmd(cmd *aranyagopb.Cmd)

	// PostMsg upload command execution result to broker/server
	PostMsg(sid uint64, kind aranyagopb.MsgType, msg proto.Marshaler) error

	PostData(sid uint64, kind aranyagopb.MsgType, seq uint64, completed bool, data []byte) (lastSeq uint64, _ error)
}
