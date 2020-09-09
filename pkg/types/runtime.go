package types

import (
	"context"
	"io"

	"arhat.dev/aranya-proto/aranyagopb"
)

// nolint:lll
type Runtime interface {
	// Name of the runtime name
	Name() string

	// Version the runtime version
	Version() string

	// OS the kernel name of the container runtime
	OS() string

	// Arch the cpu arch of the container runtime
	Arch() string

	// KernelVersion of the container runtime
	KernelVersion() string

	// SetContext used to set container runtime's context to agent's context
	SetContext(ctx context.Context)

	// InitRuntime MUST be called right after runtime has been created to start
	// all existing pods, if abbot container exists, start it first
	//
	// only fatal error will be returned
	InitRuntime() error

	// ExecInContainer execute a command in a running container
	ExecInContainer(podUID, container string, stdin io.Reader, stdout, stderr io.Writer, resizeCh <-chan *aranyagopb.TtyResizeOptions, command []string, tty bool) *aranyagopb.Error

	// AttachContainer to attach a running container's stdin/stdout/stderr
	AttachContainer(podUID, container string, stdin io.Reader, stdout, stderr io.Writer, resizeCh <-chan *aranyagopb.TtyResizeOptions) error

	// GetContainerLogs retrieve
	GetContainerLogs(podUID string, options *aranyagopb.LogOptions, stdout, stderr io.WriteCloser, logCtx context.Context) error

	// PortForward establish temporary tcp reverse proxy to cloud
	PortForward(podUID string, protocol string, port int32, downstream io.ReadWriter) error
}
