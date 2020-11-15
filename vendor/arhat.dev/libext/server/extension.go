/*
Copyright 2020 The arhat.dev Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"arhat.dev/arhat-proto/arhatgopb"
	"arhat.dev/libext/codec"
	"arhat.dev/pkg/log"
	"arhat.dev/pkg/queue"
	"golang.org/x/sync/errgroup"
)

func NewExtensionContext(
	parent context.Context,
	name string,
	codec codec.Interface,
	sendCmd CmdSendFunc,
) *ExtensionContext {
	extCtx, cancel := context.WithCancel(parent)
	return &ExtensionContext{
		Context: extCtx,
		Name:    name,
		SendCmd: sendCmd,
		Codec:   codec,

		close: cancel,
	}
}

// ExtensionContext of one extension connection
type ExtensionContext struct {
	Context context.Context
	Name    string
	SendCmd CmdSendFunc
	Codec   codec.Interface

	close context.CancelFunc
}

// Close extension connection
func (c *ExtensionContext) Close() {
	c.close()
}

type (
	// Stateful command send func, used to send command to and receive correspond
	// message from the connected extension
	CmdSendFunc func(cmd *arhatgopb.Cmd, waitForResponse bool) (*arhatgopb.Msg, error)

	// Handle message received from extension without previous command request,
	// all messages passed to this function belongs to one extension
	//
	// usually instance of this function type is used to handle extension events
	// 		e.g. lights turned off manually and detected by the extension
	OutOfBandMsgHandleFunc func(msg *arhatgopb.Msg)

	// Handle func for your own business logic
	ExtensionHandleFunc func(c *ExtensionContext)

	// Func factory to create extension specific handle func
	ExtensionHandleFuncFactory func(extensionName string) (ExtensionHandleFunc, OutOfBandMsgHandleFunc)
)

func NewExtensionManager(
	ctx context.Context,
	logger log.Interface,
	handleFuncFactory ExtensionHandleFuncFactory,
) *ExtensionManager {
	if logger == nil {
		logger = log.NoOpLogger
	}

	return &ExtensionManager{
		ctx:    ctx,
		logger: logger,

		createHandleFunc: handleFuncFactory,

		registeredExtensions: make(map[string]struct{}),

		mu: new(sync.RWMutex),
	}
}

type ExtensionManager struct {
	ctx    context.Context
	logger log.Interface

	createHandleFunc ExtensionHandleFuncFactory

	registeredExtensions map[string]struct{}

	mu *sync.RWMutex
}

// nolint:gocyclo
func (m *ExtensionManager) HandleStream(
	name string,
	codec codec.Interface,
	keepaliveInterval time.Duration,
	messageTimeout time.Duration,
	stream io.ReadWriter,
) error {
	// register new extension
	err := func() error {
		m.mu.Lock()
		defer m.mu.Unlock()

		_, exists := m.registeredExtensions[name]
		if exists {
			return fmt.Errorf("extension %q already registered", name)
		}

		m.registeredExtensions[name] = struct{}{}
		return nil
	}()
	if err != nil {
		return fmt.Errorf("failed to register extension: %w", err)
	}

	wg, ctx := errgroup.WithContext(m.ctx)

	// timeout queue for slow responses and keepalive
	tq := queue.NewTimeoutQueue()
	tq.Start(ctx.Done())

	defer func() {
		m.mu.Lock()
		defer m.mu.Unlock()

		delete(m.registeredExtensions, name)
	}()

	cmdWriteCh := make(chan *arhatgopb.Cmd, 1)
	keepaliveCh := make(chan struct{}, 1)
	wg.Go(func() error {
		enc := codec.NewEncoder(stream)
		for {
			select {
			case _, more := <-keepaliveCh:
				if !more {
					return io.ErrUnexpectedEOF
				}

				err2 := enc.Encode(&arhatgopb.Cmd{
					Kind: arhatgopb.CMD_PING,
				})
				if err2 != nil {
					// failed to ping, connection probably gone (especially for udp)
					return io.ErrUnexpectedEOF
				}
			case cmd, more := <-cmdWriteCh:
				if !more {
					return io.EOF
				}

				err2 := enc.Encode(cmd)
				if err2 != nil {
					return fmt.Errorf("failed to encode extension command: %w", err2)
				}
			}
		}
	})

	msgWait := new(sync.Map)
	sendCmd := func(cmd *arhatgopb.Cmd, waitForResponse bool) (*arhatgopb.Msg, error) {
		var waitV *msgWaitValue
		// prepare for message response wait
		if waitForResponse {
			waitV = newMsgWaitValue()
			key := msgWaitKey{
				id:  cmd.Id,
				seq: cmd.Seq,
			}
			_, loaded := msgWait.LoadOrStore(key, waitV)
			if loaded {
				waitV.close()
				return nil, fmt.Errorf("cmd sent before, no response yet")
			}
			_ = tq.OfferWithDelay(key, nil, messageTimeout)
			defer func() {
				tq.Remove(key)

				waitV.close()
				msgWait.Delete(key)
			}()
		}
		// send cmd
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case cmdWriteCh <- cmd:
		}

		// wait for message response if required
		if waitForResponse {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case msg, more := <-waitV.msgCh:
				if !more {
					return nil, fmt.Errorf("timeout")
				}

				return msg, nil
			}
		}

		return nil, nil
	}

	wg.Go(func() error {
		defer func() {
			close(keepaliveCh)
			tq.Clear()
		}()

		_ = tq.OfferWithDelay(name, nil, keepaliveInterval)
		closeMsgWait := func(v interface{}) {
			if v == nil {
				return
			}
			mv, ok := v.(*msgWaitValue)
			if !ok {
				return
			}

			mv.close()
		}

		takeCh := tq.TakeCh()
		for {
			select {
			case <-ctx.Done():
				// make everything timeout
				msgWait.Range(func(key, v interface{}) bool {
					closeMsgWait(v)
					msgWait.Delete(key)
					return true
				})
				return io.EOF
			case td := <-takeCh:
				switch td.Key.(type) {
				case msgWaitKey:
					v, _ := msgWait.Load(td.Key)
					closeMsgWait(v)
					msgWait.Delete(td.Key)
				case string:
					select {
					case keepaliveCh <- struct{}{}:
					case <-ctx.Done():
						// hand over to outer select
						continue
					}

					_ = tq.OfferWithDelay(name, nil, keepaliveInterval)
				}
			}
		}
	})

	handle, onOutOfBandMsgRecv := m.createHandleFunc(name)

	onMsgRecv := func(msg *arhatgopb.Msg) {
		key := msgWaitKey{
			id:  msg.Id,
			seq: msg.Ack,
		}

		v, ok := msgWait.Load(key)
		if !ok || v == nil {
			onOutOfBandMsgRecv(msg)
			return
		}

		mv, ok := v.(*msgWaitValue)
		if !ok {
			onOutOfBandMsgRecv(msg)
			return
		}

		mv.send(ctx.Done(), msg)
	}

	wg.Go(func() error {
		handle(NewExtensionContext(ctx, name, codec, sendCmd))

		// return trivial error to trigger context exit
		return io.EOF
	})

	// receive messages from extension handler
	wg.Go(func() error {
		dec := codec.NewDecoder(stream)
		for {
			msg := new(arhatgopb.Msg)
			err2 := dec.Decode(msg)
			if err2 != nil {
				if err2 != io.EOF {
					return fmt.Errorf("failed to decode extension message: %w", err2)
				}

				return io.EOF
			}

			select {
			case <-ctx.Done():
				return io.EOF
			default:
				onMsgRecv(msg)
			}
		}
	})

	err = wg.Wait()
	if err != nil && err != io.EOF {
		return err
	}

	return nil
}
