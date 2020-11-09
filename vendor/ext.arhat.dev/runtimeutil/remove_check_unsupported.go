// +build plan9

package runtimeutil

import "arhat.dev/pkg/wellknownerrors"

func IsLikelyNotMountPoint(file string) (bool, error) {
	return false, wellknownerrors.ErrNotSupported
}
