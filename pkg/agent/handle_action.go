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
	"syscall"
	"time"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/libext/types"
	"arhat.dev/pkg/exechelper"
	"arhat.dev/pkg/iohelper"
	"arhat.dev/pkg/nethelper"
	"arhat.dev/pkg/wellknownerrors"
	"ext.arhat.dev/runtimeutil/actionutil"

	"arhat.dev/arhat/pkg/constant"
	"arhat.dev/arhat/pkg/exec"
	"arhat.dev/arhat/pkg/util/errconv"
)

func (b *Agent) handleExec(sid uint64, data []byte) {
	opts := new(aranyagopb.ExecOrAttachCmd)

	err := opts.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal ContainerExecOrAttachCmd: %w", err))
		return
	}

	b.processInNewGoroutine(sid, "exec", func() {
		var (
			wg  = new(sync.WaitGroup)
			seq uint64
		)
		b.handleTerminalStreams(
			sid, &seq, opts.Stdout, opts.Stderr, opts.Tty, wg,
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
					cmd exec.Cmd
					err error
				)

				if opts.Stdin {
					err = b.streams.Add(sid, func() (io.WriteCloser, types.ResizeHandleFunc, error) {
						var (
							procStdin io.ReadCloser
							procInput io.WriteCloser

							// create stdin only when tty not required
							createStdin = !opts.Tty
						)

						if createStdin {
							procStdin, procInput = iohelper.Pipe()
						}

						cmd, err = exec.DoIfTryFailed(
							procStdin,
							stdout,
							stderr,
							opts.Command,
							opts.Tty,
							opts.Envs,
							false,
						)
						if err != nil {
							if procStdin != nil {
								_ = procStdin.Close()
							}
							if procInput != nil {
								_ = procInput.Close()
							}
							return nil, nil, err
						}

						if startedCmd, ok := cmd.(*exechelper.Cmd); ok && opts.Tty {
							if startedCmd.TtyInput == nil || startedCmd.TtyOutput == nil {
								_ = startedCmd.Release()
								return nil, nil, fmt.Errorf("invalid return value of cmd with tty")
							}

							// tty will create a stdin pipe, reuse it
							procInput = startedCmd.TtyInput

							// upload tty output
							wg.Add(1)
							go func() {
								defer func() {
									_ = startedCmd.TtyOutput.Close()
									wg.Done()
								}()

								b.uploadDataOutput(
									sid,
									startedCmd.TtyOutput,
									aranyagopb.MSG_DATA_STDOUT,
									&seq,
								)
							}()
						}

						return &flexWriteCloser{
								Writer: procInput,
								closeFunc: func() error {
									// close stdin with delay
									closeWithDelay(procStdin, 5*time.Second, 128*1024)
									// close input to stdin immediately
									return procInput.Close()
								},
							}, func(cols, rows uint32) {
								_ = cmd.Resize(cols, rows)
							}, nil
					})
				} else {
					cmd, err = exec.DoIfTryFailed(
						nil, stdout, stderr, opts.Command, opts.Tty, opts.Envs, false,
					)
				}

				if err != nil {
					return &aranyagopb.ErrorMsg{
						Kind:        aranyagopb.ERR_COMMON,
						Description: err.Error(),
						Code:        exechelper.DefaultExitCodeOnError,
					}
				}

				// mark stream prepared (can be obsolute)
				_, err = b.PostData(
					sid, aranyagopb.MSG_STREAM_CONTINUE, nextSeq(&seq), false, nil,
				)

				if err != nil {
					_ = cmd.Release()

					b.handleConnectivityError(sid, err)
					return &aranyagopb.ErrorMsg{
						Kind:        aranyagopb.ERR_COMMON,
						Description: err.Error(),
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
	opts := new(aranyagopb.ExecOrAttachCmd)

	err := opts.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal ContainerExecOrAttachCmd: %w", err))
		return
	}

	b.processInNewGoroutine(sid, "attach", func() {
		var (
			wg  = new(sync.WaitGroup)
			seq uint64
		)
		b.handleTerminalStreams(
			sid, &seq, opts.Stdout, opts.Stderr, true, wg,
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
						shell = "sh"
					}
				}

				var cmd *exechelper.Cmd
				err = b.streams.Add(sid, func() (io.WriteCloser, types.ResizeHandleFunc, error) {
					cmd, err = exechelper.Do(exechelper.Spec{
						Context: nil,
						Command: []string{shell},
						Stdin:   nil,
						Stdout:  nil,
						Stderr:  nil,
						Tty:     true,
					})
					if err != nil {
						return nil, nil, err
					}

					if cmd.TtyInput == nil || cmd.TtyOutput == nil {
						_ = cmd.Release()
						return nil, nil, fmt.Errorf("invalid return value of cmd with tty")
					}

					// upload tty output
					wg.Add(1)
					go func() {
						defer func() {
							_ = cmd.TtyOutput.Close()
							wg.Done()
						}()

						b.uploadDataOutput(
							sid, cmd.TtyOutput,
							aranyagopb.MSG_DATA_STDOUT,
							&seq,
						)
					}()

					return &flexWriteCloser{
							Writer: cmd.TtyInput,
							closeFunc: func() error {
								closeWithDelay(cmd.TtyOutput, 5*time.Second, 128*1024)
								return cmd.TtyInput.Close()
							},
						}, func(cols, rows uint32) {
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

				// mark stream prepared (can be obsolute)
				_, err = b.PostData(
					sid, aranyagopb.MSG_STREAM_CONTINUE, nextSeq(&seq), false, nil,
				)

				if err != nil {
					_ = cmd.Release()

					b.handleConnectivityError(sid, err)
					return &aranyagopb.ErrorMsg{
						Kind:        aranyagopb.ERR_COMMON,
						Description: err.Error(),
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
		var (
			wg  = new(sync.WaitGroup)
			seq uint64
		)
		b.handleTerminalStreams(
			sid, &seq, true, true, false, wg,
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
	io.Writer
	closeFunc func() error
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
			seq        uint64
			downstream io.ReadCloser
			closeWrite func()
			errCh      <-chan error

			pr, pw = iohelper.Pipe()
		)

		defer func() {
			_ = pw.Close()
			closeWithDelay(pr, 5*time.Second, 64*1024)
			closeWithDelay(downstream, 5*time.Second, 64*1024)

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

		err = b.streams.Add(sid, func() (io.WriteCloser, types.ResizeHandleFunc, error) {
			address := opts.Address

			if opts.Port > 0 {
				// ip based network protocols with port option
				if len(address) == 0 {
					address = "localhost"
				}

				address = net.JoinHostPort(address, strconv.FormatInt(int64(opts.Port), 10))
			}

			downstream, closeWrite, errCh, err = nethelper.Forward(
				b.ctx,
				nil,
				opts.Network,
				address,
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
				Writer: pw,
				closeFunc: func() error {
					closeWrite()

					// assume 64KB/s
					closeWithDelay(pr, 10*time.Second, 64*1024)
					closeWithDelay(downstream, 10*time.Second, 64*1024)
					return nil
				},
			}, nil, nil
		})
		if err != nil {
			return
		}

		// mark stream prepared (can be obsolute)
		_, err = b.PostData(
			sid, aranyagopb.MSG_STREAM_CONTINUE, nextSeq(&seq), false, nil,
		)
		if err != nil {
			b.handleConnectivityError(sid, err)
		}

		go func() {
			defer func() {
				_, _ = b.PostData(sid, aranyagopb.MSG_DATA, nextSeq(&seq), true, nil)
			}()

			b.uploadDataOutput(
				sid,
				downstream,
				aranyagopb.MSG_DATA,
				&seq,
			)
		}()

		// wait until downstream read exited

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

func (b *Agent) handleTerminalStreams(
	sid uint64,
	pSeq *uint64,
	useStdout, useStderr, useTty bool,
	wg *sync.WaitGroup,
	preRun func() error,
	run func(stdout, stderr io.WriteCloser) *aranyagopb.ErrorMsg,
) {
	var (
		preRunErr error
		err       *aranyagopb.ErrorMsg
	)

	defer func() {
		if err != nil {
			data, _ := err.Marshal()
			_, _ = b.PostData(sid, aranyagopb.MSG_ERROR, nextSeq(pSeq), true, data)
		} else {
			_, _ = b.PostData(sid, aranyagopb.MSG_DATA, nextSeq(pSeq), true, nil)
		}

		b.streams.Del(sid)
	}()

	if preRun != nil {
		preRunErr = preRun()
		if preRunErr != nil {
			return
		}
	}

	stdout, stderr, closeStreams := b.createStreams(
		sid, useStdout, useStderr, useTty, pSeq, wg,
	)

	err = run(stdout, stderr)

	// once run finished, doesn't mean we can close stream(s) immediately
	// because exec.Cmd won't close stdout/stderr since they are treated as
	// io.Writer, so we need to close stdout and/or stderr manually
	closeStreams()

	// and due to os.Pipe buffering, wait until stream are closed
	wg.Wait()
}

func (b *Agent) createStreams(
	sid uint64,
	useStdout, useStderr, tty bool,
	pSeq *uint64,
	wg *sync.WaitGroup,
) (stdout, stderr io.WriteCloser, close func()) {
	var (
		readStdout io.ReadCloser
		readStderr io.ReadCloser
	)

	// stdout/stderr is not used when tty is required
	if tty {
		return nil, nil, func() {}
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
				sid, readStdout, aranyagopb.MSG_DATA_STDOUT, pSeq,
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
				sid, readStderr, aranyagopb.MSG_DATA_STDERR, pSeq,
			)
		}()
	}

	return stdout, stderr, func() {
		if stdout != nil {
			_ = stdout.Close()
			closeWithDelay(readStdout, 5*time.Second, 128*1024)
		}

		if stderr != nil {
			_ = stderr.Close()
			closeWithDelay(readStderr, 5*time.Second, 128*1024)
		}
	}
}

func closeWithDelay(r io.Closer, waitAtLeast time.Duration, throughput int) {
	if r == nil {
		return
	}

	var (
		fd    uintptr
		hasFd = false
	)
	switch t := r.(type) {
	case interface {
		Fd() uintptr
	}:
		fd = t.Fd()
		hasFd = true
	case interface {
		SyscallConn() (syscall.RawConn, error)
	}:
		rawConn, err := t.SyscallConn()
		if err == nil {
			err = rawConn.Control(func(_fd uintptr) {
				fd = _fd
			})
		}
		hasFd = err == nil
	}

	// do not block close function call

	if !hasFd {
		go func() {
			time.Sleep(waitAtLeast)
			_ = r.Close()
		}()
		return
	}

	go func() {
		defer func() {
			_ = r.Close()
		}()

		var wait time.Duration
		for {
			n, err := iohelper.CheckBytesToRead(fd)
			if err != nil {
				// unable to check bytes to read
				if waitAtLeast > 0 {
					time.Sleep(waitAtLeast)
				}

				return
			}

			if n == 0 {
				// no data unread
				return
			}

			// get time to wait
			wait = time.Duration(n/throughput) * time.Second
			if wait < time.Second {
				wait = time.Second
			}

			waitAtLeast -= wait
			time.Sleep(wait)
		}
	}()
}

func (b *Agent) uploadDataOutput(
	sid uint64,
	rd io.Reader,
	kind aranyagopb.MsgType,
	pSeq *uint64,
) {
	size := b.GetClient().MaxPayloadSize()
	if size > 64*1024 {
		size = 64 * 1024
	}

	buf := make([]byte, size)
	for {
		n, err := rd.Read(buf)
		if err != nil {
			if n == 0 {
				return
			}
		}

		if n == 0 {
			continue
		}

		data := make([]byte, n)
		_ = copy(data, buf)

		// do not check returned last seq since we have limited the buffer size
		_, err = b.PostData(sid, kind, nextSeq(pSeq), false, data)
		if err != nil {
			return
		}
	}
}

func nextSeq(p *uint64) uint64 {
	return atomic.AddUint64(p, 1) - 1
}
