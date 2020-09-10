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

func (b *Agent) handlePodOperationCmd(sid uint64, data []byte) {
	cmd := new(aranyagopb.PodOperationCmd)

	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal pod cmd: %w", err))
		return
	}

	switch cmd.Action {
	case aranyagopb.PORT_FORWARD_TO_CONTAINER:
		pfOpt := cmd.GetPortForwardOptions()
		if pfOpt == nil {
			b.handleRuntimeError(sid, errRequiredOptionsNotFound)
			return
		}

		s := b.streams.NewStream(b.ctx, sid, true, false)

		b.processInNewGoroutine(sid, "ctr.port-forward", func() {
			b.doPortForward(sid, cmd.GetPortForwardOptions(), s.Reader())
		})
	case aranyagopb.EXEC_IN_CONTAINER:
		execOpt := cmd.GetExecOptions()
		if execOpt == nil {
			b.handleRuntimeError(sid, errRequiredOptionsNotFound)
			return
		}

		s := b.streams.NewStream(b.ctx, sid, execOpt.Stdin, execOpt.Tty)

		b.processInNewGoroutine(sid, "ctr.exec", func() {
			b.doContainerExec(sid, execOpt, s.Reader(), s.ResizeCh())
		})
	case aranyagopb.ATTACH_TO_CONTAINER:
		attachOpt := cmd.GetExecOptions()
		if attachOpt == nil {
			b.handleRuntimeError(sid, errRequiredOptionsNotFound)
			return
		}

		s := b.streams.NewStream(b.ctx, sid, attachOpt.Stdin, attachOpt.Tty)

		b.processInNewGoroutine(sid, "ctr.attach", func() {
			b.doContainerAttach(sid, attachOpt, s.Reader(), s.ResizeCh())
		})
	case aranyagopb.RETRIEVE_CONTAINER_LOG:
		logOpt := cmd.GetLogOptions()
		if logOpt == nil {
			b.handleRuntimeError(sid, errRequiredOptionsNotFound)
			return
		}

		s := b.streams.NewStream(b.ctx, sid, false, false)

		b.processInNewGoroutine(sid, "ctr.log", func() {
			b.doContainerLog(sid, logOpt, s.Context())
		})
	case aranyagopb.WRITE_TO_CONTAINER:
		inputOpt := cmd.GetInputOptions()
		if inputOpt == nil {
			b.handleRuntimeError(sid, errRequiredOptionsNotFound)
			return
		}

		if inputOpt.Completed {
			b.streams.CloseRead(sid, inputOpt.Seq)
			return
		}

		if !b.streams.Write(sid, inputOpt.Seq, inputOpt.Data) {
			b.handleRuntimeError(sid, errStreamSessionClosed)
			return
		}
	case aranyagopb.RESIZE_CONTAINER_TTY:
		resizeOpt := cmd.GetResizeOptions()
		if resizeOpt == nil {
			b.handleRuntimeError(sid, errRequiredOptionsNotFound)
			return
		}

		if !b.streams.Resize(sid, resizeOpt) {
			b.handleRuntimeError(sid, errStreamSessionClosed)
			return
		}
	default:
		b.handleUnknownCmd(sid, "podOperation", cmd)
		return
	}
}

func (b *Agent) handleStreamOperation(
	sid uint64, useStdin, useStdout, useStderr, useTty bool,
	preRun func() error,
	run func(stdout, stderr io.WriteCloser) *aranyagopb.Error,
) {
	var (
		seq       uint64
		preRunErr error
		err       *aranyagopb.Error
		wg        = new(sync.WaitGroup)
	)

	defer func() {
		wg.Wait()

		if err != nil {
			_ = b.PostMsg(aranyagopb.NewDataErrorMsg(sid, true, nextSeq(&seq), err))
		} else {
			_ = b.PostMsg(aranyagopb.NewDataMsg(sid, true, aranyagopb.DATA_OTHER, nextSeq(&seq), nil))
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

func (b *Agent) doContainerAttach(
	sid uint64,
	options *aranyagopb.ExecOptions,
	stdin io.Reader,
	resizeCh <-chan *aranyagopb.TtyResizeOptions,
) {
	b.handleStreamOperation(sid, options.Stdin, options.Stdout, options.Stderr, options.Tty || options.PodUid == "",
		// preRun check
		nil,
		// run
		func(stdout, stderr io.WriteCloser) *aranyagopb.Error {
			if !options.Stdin {
				stdin = nil
			}

			// container attach
			if options.PodUid != "" {
				return errconv.ToConnectivityError(
					b.runtime.AttachContainer(
						options.PodUid, options.Container,
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
				return aranyagopb.NewCommonErrorWithCode(int64(exitCode), err.Error())
			}

			return nil
		},
	)
}

func (b *Agent) doContainerExec(
	sid uint64,
	options *aranyagopb.ExecOptions,
	stdin io.Reader,
	resizeCh <-chan *aranyagopb.TtyResizeOptions,
) {
	b.handleStreamOperation(sid, options.Stdin, options.Stdout, options.Stderr, options.Tty,
		// preRun check
		func() error {
			if len(options.Command) == 0 {
				return errCommandNotProvided
			}
			return nil
		},
		// run
		func(stdout, stderr io.WriteCloser) *aranyagopb.Error {
			if !options.Stdin {
				stdin = nil
			}

			// container exec
			if options.PodUid != "" {
				return b.runtime.ExecInContainer(
					options.PodUid, options.Container,
					stdin, stdout, stderr,
					resizeCh, options.Command, options.Tty,
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
				options.Command, options.Tty, options.Envs,
			)
			if err != nil {
				return aranyagopb.NewCommonErrorWithCode(int64(exitCode), err.Error())
			}
			return nil
		},
	)
}

// nolint:golint
func (b *Agent) doContainerLog(
	sid uint64,
	options *aranyagopb.LogOptions,
	logCtx context.Context,
) {
	b.handleStreamOperation(sid, false, true, true, false,
		// preRun check
		nil,
		// run
		func(stdout, stderr io.WriteCloser) *aranyagopb.Error {
			// handle /containerLogs
			if options.PodUid != "" {
				return errconv.ToConnectivityError(
					b.runtime.GetContainerLogs(
						options.PodUid,
						options,
						stdout, stderr,
						logCtx,
					),
				)
			}

			// handle host log
			if !b.hostConfig.AllowLog {
				return errconv.ToConnectivityError(wellknownerrors.ErrNotSupported)
			}

			// handle /logs
			if options.Path != "" {
				info, err := os.Stat(options.Path)
				if err != nil {
					return errconv.ToConnectivityError(err)
				}

				if info.IsDir() {
					var files []os.FileInfo
					files, err = ioutil.ReadDir(options.Path)
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

				data, err := ioutil.ReadFile(options.Path)
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

			if options.Previous {
				file = constant.PrevLogFile(file)
			}

			err := runtimeutil.ReadLogs(logCtx, file, options, stdout, stderr)
			if err != nil {
				return errconv.ToConnectivityError(err)
			}

			return nil
		})
}

func (b *Agent) doPortForward(sid uint64, options *aranyagopb.PortForwardOptions, input io.Reader) {
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
		n, err2 := io.Copy(upstream, input)
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
				_ = b.PostMsg(aranyagopb.NewDataErrorMsg(sid, true, nextSeq(&seq), errconv.ToConnectivityError(err)))
			} else {
				_ = b.PostMsg(aranyagopb.NewDataMsg(sid, true, aranyagopb.DATA_OTHER, nextSeq(&seq), nil))
			}
		}()

		// pipe upstream received data to kubectl
		b.uploadDataOutput(sid, upstream, aranyagopb.DATA_OTHER, constant.DefaultPortForwardStreamReadTimeout, &seq)
	}()

	protocol := options.Protocol
	if protocol == "" {
		protocol = "tcp"
	}

	// container port-forward
	if options.PodUid != "" {
		err = b.runtime.PortForward(options.PodUid, protocol, options.Port, downstream)
		return
	}

	// host port-forward
	if !b.hostConfig.AllowPortForward {
		err = wellknownerrors.ErrNotSupported
		return
	}

	err = runtimeutil.PortForward(b.ctx, "localhost", protocol, options.Port, downstream)
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
	)

	if interactive {
		readTimeout = constant.DefaultInteractiveStreamReadTimeout
	}

	if useStdout {
		readStdout, stdout = iohelper.Pipe()
		wg.Add(1)
		go func() {
			defer func() {
				_ = readStdout.Close()
				wg.Done()
			}()

			b.uploadDataOutput(sid, readStdout, aranyagopb.DATA_STDOUT, readTimeout, pSeq)
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

			b.uploadDataOutput(sid, readStderr, aranyagopb.DATA_STDERR, readTimeout, pSeq)
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
	kind aranyagopb.Data_Kind,
	readTimeout time.Duration,
	pSeq *uint64,
) {
	r := iohelper.NewTimeoutReader(rd, b.GetClient().MaxDataSize())
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

		err := b.PostMsg(aranyagopb.NewDataMsg(sid, false, kind, nextSeq(pSeq), data))
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
