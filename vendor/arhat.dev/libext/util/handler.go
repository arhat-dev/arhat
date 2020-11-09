package util

import (
	"errors"
	"fmt"
	"sync"

	"arhat.dev/arhat-proto/arhatgopb"

	"arhat.dev/libext/types"
)

var ErrMsgSendFuncNotSet = errors.New("message send func not set")

func NewBaseHandler(mu *sync.RWMutex) *BaseHandler {
	return &BaseHandler{
		mu: mu,
	}
}

type BaseHandler struct {
	msgSendFunc types.MsgSendFunc

	mu *sync.RWMutex
}

func (h *BaseHandler) SetMsgSendFunc(sendMsg types.MsgSendFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.msgSendFunc = sendMsg
}

func (h *BaseHandler) SendMsg(msg *arhatgopb.Msg) error {
	h.mu.RLock()
	sendMsg := h.msgSendFunc
	h.mu.RUnlock()

	if sendMsg == nil {
		return fmt.Errorf("message send func not set")
	}

	return sendMsg(msg)
}
