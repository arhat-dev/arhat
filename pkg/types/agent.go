package types

import (
	"context"

	"arhat.dev/aranya-proto/aranyagopb"
	"github.com/gogo/protobuf/proto"
)

type (
	ConnectivityConfigFactoryFunc func() interface{}
	ConnectivityClientFactoryFunc func(agent Agent, clientConfig interface{}) (ConnectivityClient, error)
)

type RawCmdHandleFunc func(sid uint64, data []byte)

type (
	CmdHandleFunc func(cmd *aranyagopb.Cmd)
	DataPostFunc  func(sid uint64, kind aranyagopb.Kind, seq uint64, completed bool, data []byte) (uint64, error)
	MsgPostFunc   func(sid uint64, kind aranyagopb.Kind, msg proto.Marshaler) error
)

type Agent interface {
	// Context of the agent
	Context() context.Context

	// HandleCmd received from aranya
	HandleCmd(cmd *aranyagopb.Cmd)

	// PostMsg upload command execution result to broker/server
	PostMsg(sid uint64, kind aranyagopb.Kind, msg proto.Marshaler) error

	PostData(sid uint64, kind aranyagopb.Kind, seq uint64, completed bool, data []byte) (lastSeq uint64, _ error)
}

type ConnectivityClient interface {
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

	// MaxPayloadSize of a single message for this client
	MaxPayloadSize() int
}
