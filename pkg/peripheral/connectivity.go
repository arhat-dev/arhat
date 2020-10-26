// +build !noperipheral

package peripheral

import (
	"context"
	"fmt"
	"sync"

	"arhat.dev/arhat-proto/arhatgopb"
	"github.com/gogo/protobuf/proto"
)

func NewConnectivity(
	stopSig <-chan struct{},
	id uint64,
	cmdCh chan<- *arhatgopb.Cmd,
	msgCh <-chan *arhatgopb.Msg,
	onClosed func(),
) *Connectivity {
	c := &Connectivity{
		stopSig: stopSig,
		id:      id,

		cmdCh:     cmdCh,
		msgCh:     msgCh,
		expecting: make(map[uint64]chan *arhatgopb.Msg),

		onClosed: onClosed,
		mu:       new(sync.RWMutex),
	}

	go c.handleMsgs()

	return c
}

type Connectivity struct {
	stopSig <-chan struct{}
	id      uint64

	cmdCh chan<- *arhatgopb.Cmd
	msgCh <-chan *arhatgopb.Msg

	seq       uint64
	expecting map[uint64]chan *arhatgopb.Msg

	onClosed func()
	mu       *sync.RWMutex
}

func (c *Connectivity) handleMsgs() {
	for m := range c.msgCh {
		if m.Id != c.id {
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
			case <-c.stopSig:
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

func (c *Connectivity) sendCmd(ctx context.Context, p proto.Marshaler) (*arhatgopb.Msg, error) {
	seq := c.nextSeq()
	cmd, err := arhatgopb.NewCmd(c.id, seq, p)
	if err != nil {
		return nil, fmt.Errorf("failed to create peripheral cmd: %w", err)
	}

	_, ok := c.expecting[seq]
	if ok {
		return nil, fmt.Errorf("unexpected used ack")
	}

	ch := make(chan *arhatgopb.Msg, 1)
	c.mu.Lock()
	c.expecting[seq] = ch
	c.mu.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.stopSig:
		return nil, fmt.Errorf("closed")
	case c.cmdCh <- cmd:
	}

	defer func() {
		c.mu.Lock()
		delete(c.expecting, seq)
		c.mu.Unlock()
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.stopSig:
		return nil, fmt.Errorf("closed")
	case msg := <-ch:
		return msg, nil
	}
}

// Operate the peripheral via established connection
func (c *Connectivity) Operate(ctx context.Context, params map[string]string, data []byte) ([][]byte, error) {
	msg, err := c.sendCmd(ctx, &arhatgopb.PeripheralOperateCmd{
		Params: params,
		Data:   data,
	})
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
func (c *Connectivity) CollectMetrics(
	ctx context.Context, params map[string]string,
) ([]*arhatgopb.PeripheralMetricsMsg_Value, error) {
	msg, err := c.sendCmd(ctx, &arhatgopb.PeripheralMetricsCollectCmd{
		Params: params,
	})
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

func (c *Connectivity) Close() error {
	msg, err := c.sendCmd(context.TODO(), &arhatgopb.PeripheralCloseCmd{})
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
