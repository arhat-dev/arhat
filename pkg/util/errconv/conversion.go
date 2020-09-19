package errconv

import (
	"errors"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/wellknownerrors"
)

func ToConnectivityError(err error) *aranyagopb.ErrorMsg {
	if err == nil {
		return nil
	}

	var (
		msg  = err.Error()
		kind = aranyagopb.ERR_COMMON
	)

	switch {
	case errors.Is(err, wellknownerrors.ErrAlreadyExists):
		kind = aranyagopb.ERR_ALREADY_EXISTS
	case errors.Is(err, wellknownerrors.ErrNotFound):
		kind = aranyagopb.ERR_NOT_FOUND
	case errors.Is(err, wellknownerrors.ErrNotSupported):
		kind = aranyagopb.ERR_NOT_SUPPORTED
	}

	return aranyagopb.NewErrorMsg(kind, msg)
}
