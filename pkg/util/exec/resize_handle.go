package exec

import (
	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/exechelper"
)

func ResizeChannelToHandlerFunc(
	exitCh <-chan struct{},
	resizeCh <-chan *aranyagopb.ContainerTerminalResizeCmd,
) exechelper.TtyResizeSignalFunc {
	if resizeCh == nil {
		return func(doResize func(cols uint64, rows uint64) error) (more bool) {
			return false
		}
	}

	if exitCh == nil {
		return func(doResize func(cols uint64, rows uint64) error) (more bool) {
			// nolint:gosimple
			select {
			case s, more := <-resizeCh:
				if !more {
					return false
				}

				_ = doResize(uint64(s.Cols), uint64(s.Rows))
				return true
			}
		}
	}

	return func(doResize func(cols uint64, rows uint64) error) (more bool) {
		select {
		case s, more := <-resizeCh:
			if !more {
				return false
			}

			_ = doResize(uint64(s.Cols), uint64(s.Rows))
			return true
		case <-exitCh:
			return false
		}
	}
}
