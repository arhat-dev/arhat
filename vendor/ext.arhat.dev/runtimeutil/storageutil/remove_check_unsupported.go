// +build plan9

package storageutil

import "arhat.dev/pkg/wellknownerrors"

func IsLikelyNotMountPoint(file string) (bool, error) {
	return false, wellknownerrors.ErrNotSupported
}
