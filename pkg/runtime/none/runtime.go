/*
Copyright 2019 The arhat.dev Authors.

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

package none

import (
	"context"
	"io"
	"runtime"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/wellknownerrors"

	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/arhat/pkg/runtime/runtimeutil"
	"arhat.dev/arhat/pkg/types"
	"arhat.dev/arhat/pkg/util/errconv"
	"arhat.dev/arhat/pkg/util/sysinfo"
	"arhat.dev/arhat/pkg/version"
)

func NewNoneRuntime(ctx context.Context, _ types.Storage, config *conf.ArhatRuntimeConfig) (types.Runtime, error) {
	return &noneRuntime{
		BaseRuntime: runtimeutil.NewBaseRuntime(
			ctx, config, "none", version.Tag(),
			runtime.GOOS, version.Arch(), sysinfo.GetKernelVersion(),
		),
	}, nil
}

type noneRuntime struct {
	*runtimeutil.BaseRuntime
}

func (r *noneRuntime) InitRuntime() error {
	// do NOT return error!
	return nil
}

func (r *noneRuntime) ExecInContainer(
	podUID, container string,
	stdin io.Reader,
	stdout, stderr io.Writer,
	resizeCh <-chan *aranyagopb.TtyResizeOptions,
	command []string,
	tty bool,
) *aranyagopb.Error {
	return errconv.ToConnectivityError(wellknownerrors.ErrNotSupported)
}

func (r *noneRuntime) AttachContainer(
	podUID, container string,
	stdin io.Reader,
	stdout, stderr io.Writer,
	resizeCh <-chan *aranyagopb.TtyResizeOptions,
) error {
	return wellknownerrors.ErrNotSupported
}

func (r *noneRuntime) GetContainerLogs(
	podUID string,
	options *aranyagopb.LogOptions,
	stdout, stderr io.WriteCloser,
	logCtx context.Context,
) error {
	return wellknownerrors.ErrNotSupported
}

func (r *noneRuntime) PortForward(
	podUID, protocol string,
	port int32,
	upstream io.ReadWriter,
) error {
	return wellknownerrors.ErrNotSupported
}
