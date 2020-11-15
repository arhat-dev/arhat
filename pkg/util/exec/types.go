package exec

import (
	"io"
)

// tryFunc for all kinds of command execution
type tryFunc func(stdin io.Reader, stdout, stderr io.Writer, command []string, tty bool) error
