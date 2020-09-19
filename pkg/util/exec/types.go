package exec

import (
	"io"

	"arhat.dev/aranya-proto/aranyagopb"
)

// tryFunc signature for all kinds of command execution
// nolint:lll
type tryFunc func(stdin io.Reader, stdout, stderr io.Writer, resizeCh <-chan *aranyagopb.ContainerTerminalResizeCmd, command []string, tty bool) error
