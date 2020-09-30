// +build noext

package extensionutil

import "net/http"

type SyncHandleFunc func(stopSig, msgCh, cmdCh interface{}) error

type Handler struct {
	processReq SyncHandleFunc
}

func NewHandler(handler SyncHandleFunc) *Handler {
	return &Handler{}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "extension unsupported", http.StatusNotImplemented)
}
