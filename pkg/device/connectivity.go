// +build !nodev

package device

import (
	"context"
	"fmt"
	"sync"

	"arhat.dev/arhat-proto/arhatgopb"
	"github.com/gogo/protobuf/proto"
)

func NewConnectivity(
	ctx context.Context,
	id uint64,
	srv arhatgopb.DeviceExtension_SyncServer,
	msgCh <-chan *arhatgopb.DeviceMsg,
	onClosed func(),
) *Connectivity {
	c := &Connectivity{
		ctx: ctx,
		id:  id,

		msgCh:     msgCh,
		expecting: make(map[uint64]chan *arhatgopb.DeviceMsg),

		onClosed: onClosed,
		srv:      srv,
		mu:       new(sync.RWMutex),
	}

	go c.handleMsgs()

	return c
}

type Connectivity struct {
	ctx   context.Context
	id    uint64
	msgCh <-chan *arhatgopb.DeviceMsg

	seq       uint64
	expecting map[uint64]chan *arhatgopb.DeviceMsg

	onClosed func()
	srv      arhatgopb.DeviceExtension_SyncServer
	mu       *sync.RWMutex
}

func (c *Connectivity) handleMsgs() {
	for m := range c.msgCh {
		if m.DeviceId != c.id {
			// TODO: log error
			continue
		}

		if m.Ack != 0 {
			c.mu.RLock()
			ch, ok := c.expecting[m.Ack]
			c.mu.RUnlock()

			if !ok {
				// TODO: handle unexpected ack
				// discard
				continue
			}

			msg := m
			select {
			case <-c.ctx.Done():
				return
			case <-c.srv.Context().Done():
				return
			case ch <- msg:
				close(ch)
			}
		}
	}
}

func (c *Connectivity) nextSeq() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.seq++
	return c.seq
}

func (c *Connectivity) sendCmd(ctx context.Context, p proto.Marshaler) (*arhatgopb.DeviceMsg, error) {
	seq := c.nextSeq()
	cmd, err := arhatgopb.NewDeviceCmd(c.id, seq, p)
	if err != nil {
		return nil, fmt.Errorf("failed to create device cmd: %w", err)
	}

	_, ok := c.expecting[seq]
	if ok {
		return nil, fmt.Errorf("unexpected used ack")
	}

	ch := make(chan *arhatgopb.DeviceMsg, 1)
	c.mu.Lock()
	c.expecting[seq] = ch
	c.mu.Unlock()

	err = c.srv.Send(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to send device cmd: %w", err)
	}

	defer func() {
		c.mu.Lock()
		delete(c.expecting, seq)
		c.mu.Unlock()
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.ctx.Done():
		return nil, c.ctx.Err()
	case <-c.srv.Context().Done():
		return nil, c.srv.Context().Err()
	case msg := <-ch:
		return msg, nil
	}
}

// Operate the device via established connection
func (c *Connectivity) Operate(ctx context.Context, params map[string]string, data []byte) ([][]byte, error) {
	msg, err := c.sendCmd(ctx, &arhatgopb.DeviceOperateCmd{
		Params: params,
		Data:   data,
	})
	if err != nil {
		return nil, err
	}

	err = getError("failed to operate device", msg)
	if err != nil {
		return nil, err
	}

	if msg.Kind != arhatgopb.MSG_DEV_OPERATION_RESULT {
		return nil, fmt.Errorf("unexpected non operation result msg")
	}

	m := new(arhatgopb.DeviceOperateResultMsg)
	err = m.Unmarshal(msg.Payload)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to unamrshal device operation result: %w", err)
	}

	return m.Result, nil
}

// CollectMetrics collects all existing metrics for one metric kind
func (c *Connectivity) CollectMetrics(
	ctx context.Context, params map[string]string,
) ([]*arhatgopb.DeviceMetricsMsg_Value, error) {
	msg, err := c.sendCmd(ctx, &arhatgopb.DeviceMetricsCollectCmd{
		Params: params,
	})
	if err != nil {
		return nil, err
	}

	err = getError("failed to collect metrics", msg)
	if err != nil {
		return nil, err
	}

	if msg.Kind != arhatgopb.MSG_DEV_METRICS {
		return nil, fmt.Errorf("unexpected non metrics msg")
	}

	m := new(arhatgopb.DeviceMetricsMsg)
	err = m.Unmarshal(msg.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to unamrshal device metrics: %w", err)
	}

	return m.Values, nil
}

func (c *Connectivity) Close() error {
	msg, err := c.sendCmd(c.ctx, &arhatgopb.DeviceCloseCmd{})
	if err != nil {
		return err
	}

	return getError("failed to close connectivity", msg)
}

func getError(desc string, msg *arhatgopb.DeviceMsg) error {
	if msg.Kind != arhatgopb.MSG_DEV_ERROR {
		return nil
	}

	m := new(arhatgopb.ErrorMsg)
	_ = m.Unmarshal(msg.Payload)

	return fmt.Errorf("%s: %w", desc, fmt.Errorf(m.Description))
}
