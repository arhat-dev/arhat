package agent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/exechelper"
	"arhat.dev/pkg/iohelper"
	"arhat.dev/pkg/log"
	"arhat.dev/pkg/wellknownerrors"

	"arhat.dev/arhat/pkg/constant"
	"arhat.dev/arhat/pkg/runtime/runtimeutil"
	"arhat.dev/arhat/pkg/util/errconv"
	"arhat.dev/arhat/pkg/util/exec"
)

func (b *Agent) handlePodContainerExec(sid uint64, data []byte) {
	cmd := new(aranyagopb.ContainerExecOrAttachCmd)

	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal ContainerExecOrAttachCmd: %w", err))
		return
	}

	s := b.streams.NewStream(b.ctx, sid, cmd.Stdin, cmd.Tty)

	b.processInNewGoroutine(sid, "ctr.exec", func() {
		stdin := s.Reader()
		if !cmd.Stdin {
			stdin = nil
		}

		resizeCh := s.ResizeCh()

		b.handleStreamOperation(sid, cmd.Stdin, cmd.Stdout, cmd.Stderr, cmd.Tty,
			// preRun check
			func() error {
				if len(cmd.Command) == 0 {
					return errCommandNotProvided
				}
				return nil
			},
			// run
			func(stdout, stderr io.WriteCloser) *aranyagopb.ErrorMsg {
				// container exec
				if cmd.PodUid != "" {
					return b.runtime.ExecInContainer(
						cmd.PodUid, cmd.Container,
						stdin, stdout, stderr,
						resizeCh, cmd.Command, cmd.Tty,
					)
				}

				// host exec
				if !b.hostConfig.AllowExec {
					return errconv.ToConnectivityError(wellknownerrors.ErrNotSupported)
				}

				ctx, cancel := context.WithCancel(b.ctx)
				defer cancel()

				exitCode, err := exec.DoIfTryFailed(
					ctx,
					stdin, stdout, stderr,
					resizeCh,
					cmd.Command, cmd.Tty, cmd.Envs,
				)
				if err != nil {
					return aranyagopb.NewCommonErrorMsgWithCode(int64(exitCode), err.Error())
				}
				return nil
			},
		)
	})
}

func (b *Agent) handlePodContainerAttach(sid uint64, data []byte) {
	cmd := new(aranyagopb.ContainerExecOrAttachCmd)

	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal ContainerExecOrAttachCmd: %w", err))
		return
	}

	s := b.streams.NewStream(b.ctx, sid, cmd.Stdin, cmd.Tty)

	b.processInNewGoroutine(sid, "ctr.attach", func() {
		stdin := s.Reader()
		if !cmd.Stdin {
			stdin = nil
		}

		resizeCh := s.ResizeCh()

		b.handleStreamOperation(sid, cmd.Stdin, cmd.Stdout, cmd.Stderr, cmd.Tty || cmd.PodUid == "",
			// preRun check
			nil,
			// run
			func(stdout, stderr io.WriteCloser) *aranyagopb.ErrorMsg {
				if !cmd.Stdin {
					stdin = nil
				}

				// container attach
				if cmd.PodUid != "" {
					return errconv.ToConnectivityError(
						b.runtime.AttachContainer(
							cmd.PodUid, cmd.Container,
							stdin, stdout, stderr, resizeCh,
						),
					)
				}

				// host attach
				if !b.hostConfig.AllowAttach {
					return errconv.ToConnectivityError(wellknownerrors.ErrNotSupported)
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

				ctx, cancel := context.WithCancel(b.ctx)
				defer cancel()

				exitCode, err := exechelper.Do(exechelper.Spec{
					Context:        ctx,
					Command:        []string{shell},
					Stdin:          stdin,
					Stdout:         stdout,
					Stderr:         stderr,
					Tty:            true,
					OnResizeSignal: exec.ResizeChannelToHandlerFunc(ctx.Done(), resizeCh),
				})
				if err != nil {
					return aranyagopb.NewCommonErrorMsgWithCode(int64(exitCode), err.Error())
				}

				return nil
			},
		)
	})
}

func (b *Agent) handlePodContainerLogs(sid uint64, data []byte) {
	cmd := new(aranyagopb.ContainerLogsCmd)

	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal ContainerLogsCmd: %w", err))
		return
	}

	s := b.streams.NewStream(b.ctx, sid, false, false)

	b.processInNewGoroutine(sid, "ctr.logs", func() {
		b.handleStreamOperation(sid, false, true, true, false,
			// preRun check
			nil,
			// run
			func(stdout, stderr io.WriteCloser) *aranyagopb.ErrorMsg {
				// handle /containerLogs
				if cmd.PodUid != "" {
					return errconv.ToConnectivityError(
						b.runtime.GetContainerLogs(
							cmd.PodUid,
							cmd,
							stdout, stderr,
							s.Context(),
						),
					)
				}

				// handle host log
				if !b.hostConfig.AllowLog {
					return errconv.ToConnectivityError(wellknownerrors.ErrNotSupported)
				}

				// handle /logs
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

				err := runtimeutil.ReadLogs(s.Context(), file, cmd, stdout, stderr)
				if err != nil {
					return errconv.ToConnectivityError(err)
				}

				return nil
			},
		)
	})
}

func (b *Agent) handlePodPortForward(sid uint64, data []byte) {
	cmd := new(aranyagopb.PodPortForwardCmd)

	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal PodPortForwardCmd: %w", err))
		return
	}

	s := b.streams.NewStream(b.ctx, sid, true, false)

	b.processInNewGoroutine(sid, "pod.port-forward", func() {
		var (
			seq uint64
			err error

			upstream, downstream = net.Pipe()
			logger               = b.logger.WithFields(log.Uint64("sid", sid))
		)

		defer func() {
			_ = upstream.Close()

			// close this session locally (no more input data should be delivered to this session)
			b.streams.Close(sid)
		}()

		go func() {
			// pipe received remote data into upstream
			n, err2 := io.Copy(upstream, s.Reader())
			if err2 != nil && n == 0 {
				logger.I("failed to send remote data", log.Error(err2))
				return
			}

			logger.V("sent remote data", log.Int64("bytes", n), log.Error(err))
		}()

		go func() {
			defer func() {
				// send fin msg to close input in aranya
				if err != nil {
					data, _ = errconv.ToConnectivityError(err).Marshal()
					_, _ = b.PostData(sid, aranyagopb.MSG_ERROR, nextSeq(&seq), true, data)
				} else {
					_, _ = b.PostData(sid, aranyagopb.MSG_DATA_DEFAULT, nextSeq(&seq), true, nil)
				}
			}()

			// pipe upstream received data to kubectl
			b.uploadDataOutput(
				sid, upstream,
				aranyagopb.MSG_DATA_STDOUT,
				constant.DefaultPortForwardStreamReadTimeout,
				&seq, nil,
			)
		}()

		protocol := cmd.Protocol
		if protocol == "" {
			protocol = "tcp"
		}

		// container port-forward
		if cmd.PodUid != "" {
			err = b.runtime.PortForward(cmd.PodUid, protocol, cmd.Port, downstream)
			return
		}

		// host port-forward
		if !b.hostConfig.AllowPortForward {
			err = wellknownerrors.ErrNotSupported
			return
		}

		err = runtimeutil.PortForward(b.ctx, "localhost", protocol, cmd.Port, downstream)
	})
}

func (b *Agent) handlePodContainerTerminalResize(sid uint64, data []byte) {
	cmd := new(aranyagopb.ContainerTerminalResizeCmd)

	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal ContainerTerminalResizeCmd: %w", err))
		return
	}

	if !b.streams.Resize(sid, cmd) {
		b.handleRuntimeError(sid, errStreamSessionClosed)
		return
	}
}

func (b *Agent) handleStreamOperation(
	sid uint64, useStdin, useStdout, useStderr, useTty bool,
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

		b.streams.Close(sid)
	}()

	if preRun != nil {
		preRunErr = preRun()
		if preRunErr != nil {
			return
		}
	}

	stdout, stderr, closeStream := b.createTerminalStream(sid, useStdout, useStderr, useStdin && useTty, &seq, wg)
	defer closeStream()

	err = run(stdout, stderr)
}

func (b *Agent) createTerminalStream(
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

			b.uploadDataOutput(sid, readStdout, aranyagopb.MSG_DATA_STDOUT, readTimeout, pSeq, seqMu)
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

			b.uploadDataOutput(sid, readStderr, aranyagopb.MSG_DATA_STDERR, readTimeout, pSeq, seqMu)
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
	sid uint64,
	rd io.Reader,
	kind aranyagopb.Kind,
	readTimeout time.Duration,
	pSeq *uint64,
	seqMu *sync.Mutex,
) {
	r := iohelper.NewTimeoutReader(rd, b.GetClient().MaxPayloadSize())
	go r.StartBackgroundReading()

	stopSig := b.ctx.Done()
	timer := time.NewTimer(0)
	if !timer.Stop() {
		<-timer.C
	}

	defer func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}()

	for r.WaitUntilHasData(stopSig) {
		timer.Reset(readTimeout)
		data, isTimeout := r.ReadUntilTimeout(timer.C)
		if !isTimeout && !timer.Stop() {
			<-timer.C
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
