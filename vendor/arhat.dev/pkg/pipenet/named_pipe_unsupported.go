// +build plan9 js

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
	"context"
	"net"
	"os"

	"arhat.dev/pkg/wellknownerrors"
)

type PipeListener struct{}

func (c *PipeListener) Accept() (net.Conn, error) {
	return nil, wellknownerrors.ErrNotSupported
}

func (c *PipeListener) Close() error { return nil }

func (c *PipeListener) Addr() net.Addr { return &PipeAddr{} }

func ListenPipe(path, connDir string, perm os.FileMode) (net.Listener, error) {
	return nil, wellknownerrors.ErrNotSupported
}

func Dial(path string) (net.Conn, error) {
	return nil, wellknownerrors.ErrNotSupported
}

func DialContext(ctx context.Context, path string) (net.Conn, error) {
	return nil, wellknownerrors.ErrNotSupported
}

func DialPipe(laddr *PipeAddr, raddr *PipeAddr) (net.Conn, error) {
	return DialPipeContext(context.TODO(), laddr, raddr)
}

func DialPipeContext(ctx context.Context, laddr *PipeAddr, raddr *PipeAddr) (_ net.Conn, err error) {
	return nil, wellknownerrors.ErrNotSupported
}
