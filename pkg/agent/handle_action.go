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

package agent

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/libext/types"
	"arhat.dev/pkg/exechelper"
	"arhat.dev/pkg/iohelper"
	"arhat.dev/pkg/nethelper"
	"arhat.dev/pkg/wellknownerrors"
	"ext.arhat.dev/runtimeutil/actionutil"

	"arhat.dev/arhat/pkg/constant"
	"arhat.dev/arhat/pkg/util/errconv"
	"arhat.dev/arhat/pkg/util/exec"
)

func (b *Agent) handleExec(sid uint64, data []byte) {
	opts := new(aranyagopb.ExecOrAttachCmd)

	err := opts.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal ContainerExecOrAttachCmd: %w", err))
		return
	}

	b.processInNewGoroutine(sid, "exec", func() {
		b.handleStreamOperation(
			b.ctx.Done(),
			sid, opts.Stdin, opts.Stdout, opts.Stderr, opts.Tty,
			// preRun check
			func() error {
				if len(opts.Command) == 0 {
					return errCommandNotProvided
				}
				return nil
			},
			// run
			func(stdout, stderr io.WriteCloser) *aranyagopb.ErrorMsg {
				if !b.hostConfig.AllowExec {
					return &aranyagopb.ErrorMsg{
						Kind:        aranyagopb.ERR_NOT_SUPPORTED,
						Description: "host exec not allowed",
					}
				}

				var (
					cmd *exechelper.Cmd
					err error
				)

				if opts.Stdin {
					err = b.streams.Add(sid, func() (io.WriteCloser, types.ResizeHandleFunc, error) {
						pr, pw := iohelper.Pipe()
						cmd, err = exec.DoIfTryFailed(pr, stdout, stderr, opts.Command, opts.Tty, opts.Envs)
						if err != nil {
							_ = pw.Close()
							_ = pr.Close()
							return nil, nil, err
						}

						return pw, func(cols, rows uint32) {
							_ = cmd.Resize(cols, rows)
						}, nil
					})
				} else {
					cmd, err = exec.DoIfTryFailed(nil, stdout, stderr, opts.Command, opts.Tty, opts.Envs)
				}
				if err != nil {
					return &aranyagopb.ErrorMsg{
						Kind:        aranyagopb.ERR_COMMON,
						Description: err.Error(),
						Code:        exechelper.DefaultExitCodeOnError,
					}
				}

				exitCode, err := cmd.Wait()
				if err != nil {
					return &aranyagopb.ErrorMsg{
						Kind:        aranyagopb.ERR_COMMON,
						Description: err.Error(),
						Code:        int64(exitCode),
					}
				}

				return nil
			},
		)
	})
}

func (b *Agent) handleAttach(sid uint64, data []byte) {
	cmd := new(aranyagopb.ExecOrAttachCmd)

	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal ContainerExecOrAttachCmd: %w", err))
		return
	}

	b.processInNewGoroutine(sid, "attach", func() {
		b.handleStreamOperation(
			b.ctx.Done(),
			sid, cmd.Stdin, cmd.Stdout, cmd.Stderr, true,
			// preRun check
			nil,
			// run
			func(stdout, stderr io.WriteCloser) *aranyagopb.ErrorMsg {
				if !b.hostConfig.AllowAttach {
					return &aranyagopb.ErrorMsg{
						Kind:        aranyagopb.ERR_NOT_SUPPORTED,
						Description: "host attach not allowed",
					}
				}

				shell := os.Getenv("SHELL")
				if shell == "" {
					switch runtime.GOOS {
					case "windows":
						shell = "cmd"
					default:
						shell = "/bin/sh"
					}
				}

				var cmd *exechelper.Cmd
				err = b.streams.Add(sid, func() (io.WriteCloser, types.ResizeHandleFunc, error) {
					pr, pw := iohelper.Pipe()
					cmd, err = exechelper.Do(exechelper.Spec{
						Context: nil,
						Command: []string{shell},
						Stdin:   pr,
						Stdout:  stdout,
						Stderr:  stderr,
						Tty:     true,
					})
					if err != nil {
						_ = pw.Close()
						_ = pr.Close()
						return nil, nil, err
					}

					return pw, func(cols, rows uint32) {
						_ = cmd.Resize(cols, rows)
					}, nil
				})
				if err != nil {
					return &aranyagopb.ErrorMsg{
						Kind:        aranyagopb.ERR_COMMON,
						Description: err.Error(),
						Code:        exechelper.DefaultExitCodeOnError,
					}
				}

				var exitCode int
				exitCode, err = cmd.Wait()
				if err != nil {
					return &aranyagopb.ErrorMsg{
						Kind:        aranyagopb.ERR_COMMON,
						Description: err.Error(),
						Code:        int64(exitCode),
					}
				}

				return nil
			},
		)
	})
}

func (b *Agent) handleLogs(sid uint64, data []byte) {
	cmd := new(aranyagopb.LogsCmd)

	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal ContainerLogsCmd: %w", err))
		return
	}

	b.processInNewGoroutine(sid, "logs", func() {
		b.handleStreamOperation(
			b.ctx.Done(),
			sid, false, true, true, false,
			// preRun check
			nil,
			// run
			func(stdout, stderr io.WriteCloser) *aranyagopb.ErrorMsg {
				if !b.hostConfig.AllowLog {
					return &aranyagopb.ErrorMsg{
						Kind:        aranyagopb.ERR_NOT_SUPPORTED,
						Description: "host logs not allowed",
					}
				}

				if cmd.Path != "" {
					info, err := os.Stat(cmd.Path)
					if err != nil {
						return errconv.ToConnectivityError(err)
					}

					if info.IsDir() {
						var files []os.FileInfo
						files, err = ioutil.ReadDir(cmd.Path)
						if err != nil {
							return errconv.ToConnectivityError(err)
						}

						buf := new(bytes.Buffer)
						buf.WriteString(constant.IdentifierLogDir)
						buf.WriteByte('\n')
						for _, f := range files {
							buf.WriteString(filepath.Base(f.Name()))
							buf.WriteByte('\n')
						}

						_, err = buf.WriteTo(stdout)
						if err != nil {
							return errconv.ToConnectivityError(err)
						}

						return nil
					}

					data, err := ioutil.ReadFile(cmd.Path)
					if err != nil {
						return errconv.ToConnectivityError(err)
					}

					_, _ = stdout.Write([]byte(constant.IdentifierLogFile))
					_, _ = stdout.Write([]byte{'\n'})
					_, err = stdout.Write(data)
					if err != nil {
						return errconv.ToConnectivityError(err)
					}

					return nil
				}

				// handle arhat log
				file := b.kubeLogFile
				if file == "" {
					return errconv.ToConnectivityError(wellknownerrors.ErrNotFound)
				}

				if cmd.Previous {
					file = constant.PrevLogFile(file)
				}

				err := actionutil.ReadLogs(b.ctx, file, cmd, stdout, stderr)
				if err != nil {
					return errconv.ToConnectivityError(err)
				}

				return nil
			},
		)
	})
}

type flexWriteCloser struct {
	writeFunc func([]byte) (int, error)
	closeFunc func() error
}

func (a *flexWriteCloser) Write(p []byte) (n int, err error) {
	return a.writeFunc(p)
}

func (a *flexWriteCloser) Close() error {
	return a.closeFunc()
}

func (b *Agent) handlePortForward(sid uint64, data []byte) {
	opts := new(aranyagopb.PortForwardCmd)
	err := opts.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal PodPortForwardCmd: %w", err))
		return
	}

	b.processInNewGoroutine(sid, "port-forward", func() {
		var (
			seq uint64

			pr, pw = iohelper.Pipe()
		)

		defer func() {
			_ = pw.Close()

			kind := aranyagopb.MSG_DATA
			var payload []byte
			// send fin msg to close input in aranya
			if err != nil {
				kind = aranyagopb.MSG_ERROR
				payload, _ = (&aranyagopb.ErrorMsg{
					Kind:        aranyagopb.ERR_COMMON,
					Description: err.Error(),
					Code:        0,
				}).Marshal()
			}

			// best effort
			_, _ = b.PostData(sid, kind, nextSeq(&seq), true, payload)

			// close this session locally (no more input data should be delivered to this session)
			b.streams.Del(sid)
		}()

		var (
			downstream io.ReadCloser
			closeWrite func()
			errCh      <-chan error
		)

		err = b.streams.Add(sid, func() (io.WriteCloser, types.ResizeHandleFunc, error) {
			downstream, closeWrite, errCh, err = nethelper.Forward(
				b.ctx, nil, opts.Protocol,
				net.JoinHostPort("localhost", strconv.FormatInt(int64(opts.Port), 10)),
				pr,
				nil,
			)
			if err != nil {
				return nil, nil, err
			}

			if downstream == nil || closeWrite == nil || errCh == nil {
				return nil, nil, fmt.Errorf("bad port-forward implementation, missing required return values")
			}

			return &flexWriteCloser{
				writeFunc: pw.Write,
				closeFunc: func() error {
					f, ok := pr.(*os.File)
					if !ok {
						go func() {
							time.Sleep(5 * time.Second)
							_ = pw.Close()
							closeWrite()
							_ = downstream.Close()
						}()
						return nil
					}

					n, _ := iohelper.CheckBytesToRead(f.Fd())

					// assume 100 KB/s
					wait := time.Duration(n) / 1024 / 100 * time.Second
					if wait > 0 {
						go func() {
							time.Sleep(wait)
							_ = pw.Close()
							closeWrite()
							_ = downstream.Close()
						}()
					}

					return nil
				},
			}, nil, nil
		})
		if err != nil {
			return
		}

		// pipe received data to extension hub
		b.uploadDataOutput(
			b.ctx.Done(),
			sid,
			downstream,
			aranyagopb.MSG_DATA,
			20*time.Millisecond,
			&seq,
			nil,
		)

		// downstream read exited

		for {
			// drain errCh
			select {
			case <-b.ctx.Done():
				return
			case e, more := <-errCh:
				if e != nil {
					if err == nil {
						err = e
					} else {
						err = fmt.Errorf("%v; %w", err, e)
					}
				}

				if !more {
					return
				}
			}
		}
	})
}

func (b *Agent) handleTerminalResize(sid uint64, data []byte) {
	opts := new(aranyagopb.TerminalResizeCmd)
	err := opts.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal ContainerTerminalResizeCmd: %w", err))
		return
	}

	b.streams.Resize(sid, opts.Cols, opts.Rows)
}

func (b *Agent) handleStreamOperation(
	stopSig <-chan struct{},
	sid uint64,
	useStdin, useStdout, useStderr, useTty bool,
	preRun func() error,
	run func(stdout, stderr io.WriteCloser) *aranyagopb.ErrorMsg,
) {
	var (
		seq       uint64
		preRunErr error
		err       *aranyagopb.ErrorMsg
		wg        = new(sync.WaitGroup)
	)

	defer func() {
		wg.Wait()

		if err != nil {
			data, _ := err.Marshal()
			_, _ = b.PostData(sid, aranyagopb.MSG_ERROR, nextSeq(&seq), true, data)
		} else {
			_, _ = b.PostData(sid, aranyagopb.MSG_DATA_DEFAULT, nextSeq(&seq), true, nil)
		}

		b.streams.Del(sid)
	}()

	if preRun != nil {
		preRunErr = preRun()
		if preRunErr != nil {
			return
		}
	}

	stdout, stderr, closeStream := b.createTerminalStream(
		stopSig, sid, useStdout, useStderr, useStdin && useTty, &seq, wg,
	)
	defer closeStream()

	err = run(stdout, stderr)
}

func (b *Agent) createTerminalStream(
	stopSig <-chan struct{},
	sid uint64,
	useStdout, useStderr, interactive bool,
	pSeq *uint64,
	wg *sync.WaitGroup,
) (stdout, stderr io.WriteCloser, close func()) {
	var (
		readStdout  io.ReadCloser
		readStderr  io.ReadCloser
		readTimeout = constant.DefaultNonInteractiveStreamReadTimeout
		seqMu       *sync.Mutex
	)

	if interactive {
		readTimeout = constant.DefaultInteractiveStreamReadTimeout
	}

	if useStdout && useStderr {
		seqMu = new(sync.Mutex)
	}

	if useStdout {
		readStdout, stdout = iohelper.Pipe()
		wg.Add(1)
		go func() {
			defer func() {
				_ = readStdout.Close()
				wg.Done()
			}()

			b.uploadDataOutput(
				stopSig, sid, readStdout,
				aranyagopb.MSG_DATA_STDOUT,
				readTimeout, pSeq, seqMu,
			)
		}()
	}

	if useStderr {
		readStderr, stderr = iohelper.Pipe()
		wg.Add(1)
		go func() {
			defer func() {
				_ = readStderr.Close()
				wg.Done()
			}()

			b.uploadDataOutput(
				stopSig, sid, readStderr,
				aranyagopb.MSG_DATA_STDERR,
				readTimeout, pSeq, seqMu,
			)
		}()
	}

	return stdout, stderr, func() {
		if stdout != nil {
			_ = stdout.Close()
		}

		if stderr != nil {
			_ = stderr.Close()
		}
	}
}

func (b *Agent) uploadDataOutput(
	stopSig <-chan struct{},
	sid uint64,
	rd io.Reader,
	kind aranyagopb.MsgType,
	readTimeout time.Duration,
	pSeq *uint64,
	seqMu *sync.Mutex,
) {
	r := iohelper.NewTimeoutReader(rd)
	go r.FallbackReading(stopSig)

	buf := make([]byte, b.GetClient().MaxPayloadSize())
	for r.WaitForData(stopSig) {
		data, shouldCopy, err := r.Read(readTimeout, buf)
		if err != nil && err != iohelper.ErrDeadlineExceeded {
			return
		}

		if shouldCopy {
			data = make([]byte, len(data))
			_ = copy(data, buf[:len(data)])
		}

		if seqMu != nil {
			seqMu.Lock()
		}

		lastSeq, err := b.PostData(sid, kind, nextSeq(pSeq), false, data)
		atomic.StoreUint64(pSeq, lastSeq+1)

		if seqMu != nil {
			seqMu.Unlock()
		}

		if err != nil {
			b.handleConnectivityError(sid, err)
			return
		}
	}
}

func nextSeq(p *uint64) uint64 {
	seq := atomic.LoadUint64(p)
	for !atomic.CompareAndSwapUint64(p, seq, seq+1) {
		seq++
	}

	return seq
}
