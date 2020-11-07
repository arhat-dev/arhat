// +build plan9

package pipenet

import (
	"arhat.dev/pkg/wellknownerrors"
)

func mkfifo(path string, perm uint32) error {
	return wellknownerrors.ErrNotSupported
}
