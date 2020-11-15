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

	"arhat.dev/pkg/wellknownerrors"
)

func (c *ListenConfig) ListenPipe(path string) (net.Listener, error) {
	return nil, wellknownerrors.ErrNotSupported
}

func (d *Dialer) Dial(path string) (net.Conn, error) {
	return nil, wellknownerrors.ErrNotSupported
}

func (d *Dialer) DialContext(ctx context.Context, path string) (net.Conn, error) {
	return nil, wellknownerrors.ErrNotSupported
}

func (d *Dialer) DialPipe(raddr *PipeAddr) (net.Conn, error) {
	return nil, wellknownerrors.ErrNotSupported
}

func (d *Dialer) DialPipeContext(ctx context.Context, raddr *PipeAddr) (net.Conn, error) {
	return nil, wellknownerrors.ErrNotSupported
}
