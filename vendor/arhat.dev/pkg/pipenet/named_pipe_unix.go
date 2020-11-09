// +build !windows,!plan9

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

package pipenet

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"arhat.dev/pkg/hashhelper"
	"arhat.dev/pkg/iohelper"
)

var _ net.Conn = (*pipeConn)(nil)

type pipeConn struct {
	r *os.File
	w *os.File

	remoteWrite *os.File
	rawConn     syscall.RawConn
}

func (c *pipeConn) Read(b []byte) (int, error) {
	var (
		errno      syscall.Errno
		ptr        = uintptr(0)
		size       = uintptr(len(b))
		count      uintptr
		err        error
		rawConnErr error
		eC         = int(0)
	)

	if size != 0 {
		ptr = uintptr(unsafe.Pointer(&b[0]))
	}

rawRead:
	rawConnErr = c.rawConn.Read(func(fd uintptr) bool {
		for eC = 0; eC < 1024; eC++ {
			count, _, errno = syscall.Syscall(syscallRead, fd, ptr, size)
			switch {
			case errno == syscall.EAGAIN:
				// EAGAIN means resource not available, may need to wait for some time,
				// yield cpu and retry
				runtime.Gosched()
			case errno != 0:
				// error happened
				err = errno
				if count > size {
					count = 0
				}

				return true
			case count == 0:
				// no error happened and no data read, the file is closed
				err = os.ErrClosed
				return true
			default:
				err = nil
				return true
			}
		}

		count = 0
		return true
	})

	if count != 0 || err != nil {
		return int(count), err
	}

	if rawConnErr != nil {
		// this will happen when reading a closed file
		// see https://github.com/golang/go/issues/29828
		return int(count), rawConnErr
	}

	goto rawRead
}

func (c *pipeConn) Write(b []byte) (n int, err error) {
	return c.w.Write(b)
}

// Close pipe connection
func (c *pipeConn) Close() error {
	err := c.r.Close()

	_ = c.remoteWrite.Close()
	_ = os.Remove(c.r.Name())
	_ = os.Remove(c.remoteWrite.Name())

	if c.w != nil {
		err2 := c.w.Close()
		if err2 != nil {
			if err != nil {
				err = fmt.Errorf("%s; %w", err.Error(), err2)
			} else {
				err = err2
			}
		}

		_ = os.Remove(c.w.Name())
	}

	return err
}

func (c *pipeConn) LocalAddr() net.Addr {
	return &PipeAddr{Path: c.r.Name()}
}

func (c *pipeConn) RemoteAddr() net.Addr {
	return &PipeAddr{Path: c.w.Name()}
}

func (c *pipeConn) SetDeadline(t time.Time) error {
	err := c.r.SetDeadline(t)
	err2 := c.w.SetDeadline(t)
	if err2 != nil {
		if err != nil {
			err = fmt.Errorf("%s; %w", err.Error(), err2)
		} else {
			err = err2
		}
	}

	return err
}

func (c *pipeConn) SetReadDeadline(t time.Time) error {
	return c.r.SetReadDeadline(t)
}

func (c *pipeConn) SetWriteDeadline(t time.Time) error {
	return c.w.SetWriteDeadline(t)
}

var _ net.Listener = (*PipeListener)(nil)

type PipeListener struct {
	connDir string

	r      *bufio.Reader
	listen *pipeConn

	pcs    map[string]*pipeConn
	mu     *sync.RWMutex
	closed bool
}

// Accept new pipe connections, once we have received incoming coming message
// from `path`, the message content is expected to be the path of a named pipe
// listened by the client
//
// If the provided path is a valid named pipe, the listener will create another
// named pipe for client and write to the pipe provided by the client
func (c *PipeListener) Accept() (net.Conn, error) {
accept:
	path, err := c.r.ReadString('\n')
	if err != nil {
		return nil, err
	}

	path = strings.TrimRight(path, "\n")
	if path == "" {
		// new line prefix
		goto accept
	}

	if !filepath.IsAbs(path) {
		goto accept
	}

	// incoming path should be a pipe for server to write, so we can write
	// response to the client
	serverW, err := os.OpenFile(path, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		goto accept
	}

	conn, err := func() (_ *pipeConn, err error) {
		defer func() {
			if err != nil {
				_, _ = serverW.WriteString(fmt.Sprintf("error:%v\n", err))
				_ = serverW.Close()
			}

			c.mu.Unlock()
		}()
		c.mu.Lock()

		if c.closed {
			return nil, io.EOF
		}

		// close old session if any
		oldConn, exists := c.pcs[path]
		if exists {
			_ = oldConn.Close()
		}

		// open a new pipe writer for this client
		serverReadFile := filepath.Join(c.connDir, hashhelper.MD5SumHex([]byte(path)))

		var serverR, clientW *os.File
		serverR, clientW, err = createPipe(serverReadFile, 0666)
		if err != nil {
			return nil, err
		}

		defer func() {
			// always close remote writer
			_ = clientW.Close()
			if err != nil {
				_ = serverR.Close()
			}
		}()

		// notify client new read pipe
		_, err = serverW.Write(append([]byte(serverReadFile), '\n'))
		if err != nil {
			return nil, err
		}

		var rawConn syscall.RawConn
		rawConn, err = serverR.SyscallConn()
		if err != nil {
			return nil, err
		}

		conn := &pipeConn{
			r: serverR,
			w: serverW,

			remoteWrite: clientW,
			rawConn:     rawConn,
		}

		// wait for one byte acknowledge

		expected := append([]byte(hashhelper.Sha512SumHex([]byte(serverReadFile))), '\n')
		ack := make([]byte, len(expected))

		// checksum read must be finished in a single read call
		_, err = conn.Read(ack)
		if err != nil {
			return nil, err
		}
		if !bytes.Equal(expected, ack) {
			return nil, fmt.Errorf("invalid pipe path checksum")
		}

		// once we received the acknowledge, the client is expected to have the writer pipe
		// open, so we can close it and the close of the writer pipe on client side will
		// cause reader pipe error

		c.pcs[path] = conn
		return conn, nil
	}()
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (c *PipeListener) Close() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, conn := range c.pcs {
		_ = conn.Close()
	}

	return c.listen.Close()
}

func (c *PipeListener) Addr() net.Addr {
	return c.listen.LocalAddr()
}

// ListenPipe will create a named pipe at path and listen incomming message
func ListenPipe(path, connDir string, perm os.FileMode) (net.Listener, error) {
	if path == "" {
		path = os.TempDir()
	}

	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	p, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		// listen path not found, treat as pipe file to be created
	} else if p.IsDir() {
		path, err = iohelper.TempFilename(path, "pipe-listener-*")
		if err != nil {
			return nil, fmt.Errorf("failed to generate random listen addr")
		}
	}

	if connDir == "" {
		connDir, err = ioutil.TempDir(os.TempDir(), "pipe-server-*")
	} else {
		connDir, err = filepath.Abs(connDir)
	}
	if err != nil {
		return nil, err
	}

	// max path length is 4096
	// we use hex(md5sum(remote path)) as local write path
	if len(connDir) > 4096-md5.Size*2 {
		return nil, fmt.Errorf("connDir name too long")
	}

	dirInfo, err := os.Stat(connDir)
	if err != nil {
		return nil, err
	}
	if !dirInfo.IsDir() {
		return nil, fmt.Errorf("connDir is not a directory")
	}

	localR, clientReq, err := createPipe(path, uint32(perm))
	if err != nil {
		return nil, err
	}

	conn, err := localR.SyscallConn()
	if err != nil {
		return nil, err
	}

	listen := &pipeConn{
		r: localR,

		remoteWrite: clientReq,
		rawConn:     conn,
	}

	return &PipeListener{
		connDir: connDir,
		// linux path size limit is 4096
		r:      bufio.NewReaderSize(listen, 4096),
		listen: listen,
		pcs:    make(map[string]*pipeConn),
		mu:     new(sync.RWMutex),
	}, nil
}

func Dial(path string) (net.Conn, error) {
	return DialPipe(nil, &PipeAddr{Path: path})
}

func DialContext(ctx context.Context, path string) (net.Conn, error) {
	return DialPipeContext(ctx, nil, &PipeAddr{Path: path})
}

func DialPipe(laddr *PipeAddr, raddr *PipeAddr) (net.Conn, error) {
	return DialPipeContext(context.TODO(), laddr, raddr)
}

func DialPipeContext(ctx context.Context, laddr *PipeAddr, raddr *PipeAddr) (_ net.Conn, err error) {
	if raddr == nil {
		return nil, &net.AddrError{Err: "no remote address provided"}
	}

	if laddr == nil || laddr.Path == "" {
		laddr = &PipeAddr{Path: os.TempDir()}
	} else {
		var localPath string
		localPath, err = filepath.Abs(laddr.Path)
		if err != nil {
			return nil, err
		}

		laddr = &PipeAddr{Path: localPath}
	}

	// check if local path exists
	f, err := os.Stat(laddr.Path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		// not exists, treat it as file path to create pipe
	} else if f.IsDir() {
		// use a random temporary file
		laddr.Path, err = iohelper.TempFilename(laddr.Path, "pipe-client-*")
		if err != nil {
			return nil, err
		}
	}

	// create a pipe for server to write response
	clientR, serverW, err := createPipe(laddr.Path, 0666)
	if err != nil {
		return nil, err
	}

	defer func() {
		// always close serverW so we can be notified when server closed our writer
		_ = serverW.Close()
		if err != nil {
			_ = clientR.Close()
		}
	}()

	// request a new pipe from server
	clientReq, err := os.OpenFile(raddr.Path, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		return nil, err
	}

	_, err = clientReq.Write(append(append([]byte{'\n'}, []byte(laddr.Path)...), '\n'))
	_ = clientReq.Close()
	if err != nil {
		return nil, &net.OpError{Op: "dial", Net: raddr.Network(), Err: err}
	}

	dialDeadline, ok := ctx.Deadline()
	if !ok {
		// no deadline set
		dialDeadline = time.Now().Add(time.Minute)
	}

	connected := make(chan struct{})
	err = clientR.SetReadDeadline(dialDeadline)
	if err != nil {
		err = nil
		go func() {
			t := time.NewTimer(time.Until(dialDeadline))
			defer func() {
				if !t.Stop() {
					<-t.C
				}
			}()

			select {
			case <-ctx.Done():
				_ = clientR.Close()
				return
			case <-t.C:
				// timeout
				_ = clientR.Close()
				return
			case <-connected:
				_ = clientR.SetReadDeadline(time.Time{})
				return
			}
		}()
	}

	rawConn, err := clientR.SyscallConn()
	if err != nil {
		return
	}

	conn := &pipeConn{
		r:           clientR,
		remoteWrite: serverW,
		rawConn:     rawConn,
	}

	br := bufio.NewReaderSize(conn, 4096)
	serverReadFile, err := br.ReadString('\n')
	close(connected)
	if err != nil {
		return nil, &net.OpError{Op: "read", Net: raddr.Network(), Err: err}
	}

	serverReadFile = strings.TrimRight(serverReadFile, "\n")
	if strings.HasPrefix(serverReadFile, "error:") {
		return nil, &net.OpError{
			Op: "request", Net: raddr.Network(), Err: fmt.Errorf(strings.TrimPrefix(serverReadFile, "error:")),
		}
	}

	if !filepath.IsAbs(serverReadFile) {
		return nil, &net.AddrError{Addr: serverReadFile, Err: "server did not provide absolute path pipe"}
	}

	// open server provided file as writer
	conn.w, err = os.OpenFile(serverReadFile, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		return nil, err
	}

	// send filename checksum
	_, err = conn.Write(append([]byte(hashhelper.Sha512SumHex([]byte(serverReadFile))), '\n'))
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func createPipe(path string, perm uint32) (r, w *os.File, err error) {
	defer func() {
		if err == nil {
			return
		}

		if r != nil {
			_ = r.Close()
		}

		if w != nil {
			_ = w.Close()
		}

		_ = os.Remove(path)
	}()

	err = mkfifo(path, perm)
	if err != nil {
		return
	}

	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)

		var err2 error
		w, err2 = os.OpenFile(path, os.O_WRONLY, os.ModeNamedPipe)
		if err2 != nil {
			errCh <- err2
			return
		}

		_ = w.SetWriteDeadline(time.Now().Add(time.Second))
		_, err2 = w.Write([]byte{'\n'})
		_ = w.SetWriteDeadline(time.Time{})
		errCh <- err2
	}()

	// open fifo file fd as non-blocking so we can cancel read operations when required
	// see: https://github.com/golang/go/issues/20110#issuecomment-298124931
	r, err = os.OpenFile(path, os.O_RDONLY|syscall.O_CLOEXEC|syscall.O_NONBLOCK, os.ModeNamedPipe)
	if err != nil {
		return
	}

	err = <-errCh
	if err != nil {
		return
	}

	return r, w, checkInitialByte(r)
}

func checkInitialByte(fR io.Reader) error {
	initialByte := make([]byte, 1)
	_, err := fR.Read(initialByte)
	if err != nil {
		return err
	}

	if initialByte[0] != '\n' {
		return fmt.Errorf("unexpected byte %q", initialByte[0])
	}

	return nil
}
