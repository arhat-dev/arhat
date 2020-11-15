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
	"context"

	"arhat.dev/arhat-proto/arhatgopb"
)

type (
	ResizeHandleFunc func(cols, rows uint32)
	MsgSendFunc      func(msg *arhatgopb.Msg) error
)

// Handler for controller
type Handler interface {
	// SetMsgSendFunc is called by controller when new connection established, handler
	// should use the latest message send func for SendMsg function call
	SetMsgSendFunc(sendMsg MsgSendFunc)

	// SendMsg send out of band message to the extension hub
	SendMsg(msg *arhatgopb.Msg) error

	// HandleCmd process one command per function call, payload is non stream data
	HandleCmd(
		ctx context.Context,
		id, seq uint64,
		kind arhatgopb.CmdType,
		payload []byte,
	) (arhatgopb.MsgType, interface{}, error)
}
