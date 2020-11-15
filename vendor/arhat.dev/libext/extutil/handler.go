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
