package peripheral

import (
	"context"
	"fmt"
	"runtime"
	"sync/atomic"

	"arhat.dev/arhat-proto/arhatgopb"
	"arhat.dev/libext/server"
	"arhat.dev/libext/util"
	"github.com/gogo/protobuf/proto"
)

func NewConnectivity(
	id uint64,
	ec *server.ExtensionContext,
) *Conn {
	c := &Conn{
		id:      id,
		seq:     1,
		working: 0,

		sendCmd: nil,
	}

	c.sendCmd = func(ctx context.Context, kind arhatgopb.CmdType, p proto.Marshaler) (*arhatgopb.Msg, error) {
		seq := c.nextSeq()
		cmd, err := util.NewCmd(ec.Codec.Marshal, kind, id, seq, p)
		if err != nil {
			return nil, fmt.Errorf("failed to create peripheral cmd: %w", err)
		}

		return ec.SendCmd(cmd)
	}

	return c
}

type Conn struct {
	id      uint64
	seq     uint64
	working uint32

	sendCmd func(ctx context.Context, kind arhatgopb.CmdType, p proto.Marshaler) (*arhatgopb.Msg, error)
}

func (c *Conn) nextSeq() uint64 {
	defer func() {
		for !atomic.CompareAndSwapUint32(&c.working, 1, 0) {
			runtime.Gosched()
		}
	}()

	for !atomic.CompareAndSwapUint32(&c.working, 0, 1) {
		runtime.Gosched()
	}
	// TODO: reuse sequence
	c.seq++

	return c.seq
}

// Operate the peripheral via established connection
func (c *Conn) Operate(ctx context.Context, params map[string]string, data []byte) ([][]byte, error) {
	msg, err := c.sendCmd(ctx, arhatgopb.CMD_PERIPHERAL_OPERATE,
		&arhatgopb.PeripheralOperateCmd{
			Params: params,
			Data:   data,
		},
	)
	if err != nil {
		return nil, err
	}

	err = getError("failed to operate peripheral", msg)
	if err != nil {
		return nil, err
	}

	if msg.Kind != arhatgopb.MSG_PERIPHERAL_OPERATION_RESULT {
		return nil, fmt.Errorf("unexpected non operation result msg")
	}

	m := new(arhatgopb.PeripheralOperationResultMsg)
	err = m.Unmarshal(msg.Payload)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to unamrshal peripheral operation result: %w", err)
	}

	return m.Result, nil
}

// CollectMetrics collects all existing metrics for one metric kind
func (c *Conn) CollectMetrics(
	ctx context.Context, params map[string]string,
) ([]*arhatgopb.PeripheralMetricsMsg_Value, error) {
	msg, err := c.sendCmd(ctx, arhatgopb.CMD_PERIPHERAL_COLLECT_METRICS,
		&arhatgopb.PeripheralMetricsCollectCmd{
			Params: params,
		},
	)
	if err != nil {
		return nil, err
	}

	err = getError("failed to collect metrics", msg)
	if err != nil {
		return nil, err
	}

	if msg.Kind != arhatgopb.MSG_PERIPHERAL_METRICS {
		return nil, fmt.Errorf("unexpected non metrics msg")
	}

	m := new(arhatgopb.PeripheralMetricsMsg)
	err = m.Unmarshal(msg.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to unamrshal peripheral metrics: %w", err)
	}

	return m.Values, nil
}

func (c *Conn) Close() error {
	msg, err := c.sendCmd(context.TODO(), arhatgopb.CMD_PERIPHERAL_CLOSE,
		&arhatgopb.PeripheralCloseCmd{},
	)
	if err != nil {
		return err
	}

	return getError("failed to close connectivity", msg)
}

func getError(desc string, msg *arhatgopb.Msg) error {
	if msg.Kind != arhatgopb.MSG_ERROR {
		return nil
	}

	m := new(arhatgopb.ErrorMsg)
	_ = m.Unmarshal(msg.Payload)

	return fmt.Errorf("%s: %w", desc, fmt.Errorf(m.Description))
}
