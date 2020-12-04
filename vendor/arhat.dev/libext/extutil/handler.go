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

package extutil

import (
	"errors"
	"sync/atomic"

	"arhat.dev/arhat-proto/arhatgopb"

	"arhat.dev/libext/types"
)

var ErrMsgSendFuncNotSet = errors.New("message send func not set")

func NewBaseHandler() *BaseHandler {
	return &BaseHandler{
		msgSendFuncStore: new(atomic.Value),
	}
}

type BaseHandler struct {
	msgSendFuncStore *atomic.Value
}

func (h *BaseHandler) SetMsgSendFunc(sendMsg types.MsgSendFunc) {
	h.msgSendFuncStore.Store(sendMsg)
}

func (h *BaseHandler) SendMsg(msg *arhatgopb.Msg) error {
	if o := h.msgSendFuncStore.Load(); o != nil {
		if f, ok := o.(types.MsgSendFunc); ok && f != nil {
			return f(msg)
		}
	}

	return ErrMsgSendFuncNotSet
}
