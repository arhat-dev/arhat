// +build windows js plan9 aix

package exechelper

import (
	"io"
	"os/exec"

	"arhat.dev/pkg/log"
	"arhat.dev/pkg/wellknownerrors"
)

func startCmdWithTty(logger log.Interface, cmd *exec.Cmd, stdin io.Reader, stdout io.Writer, resizeH TtyResizeSignalFunc) (func(), error) {
	return nil, wellknownerrors.ErrNotSupported
}
