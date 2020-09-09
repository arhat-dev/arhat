package types

import (
	"context"

	"arhat.dev/aranya-proto/aranyagopb"
)

type Agent interface {
	// Context of the agent
	Context() context.Context

	// HandleCmd received from aranya
	HandleCmd(cmd *aranyagopb.Cmd)
}

type AgentConnectivity interface {
	// Context of this client
	Context() context.Context

	// Connect to server/broker
	Connect(dialCtx context.Context) error

	// Start internal logic to get prepared for communication with aranya
	Start(ctx context.Context) error

	// PostMsg to aranya
	PostMsg(msg *aranyagopb.Msg) error

	// Close this client
	Close() error

	// MaxDataSize of a single message for this client
	MaxDataSize() int
}
