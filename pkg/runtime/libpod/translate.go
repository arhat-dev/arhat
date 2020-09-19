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

package libpod

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"arhat.dev/aranya-proto/aranyagopb/aranyagoconst"

	"github.com/containers/libpod/v2/libpod"
	"github.com/containers/libpod/v2/libpod/define"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/log"

	"arhat.dev/arhat/pkg/constant"
)

type fakeWriterCloser struct {
	io.Writer
}

func (f *fakeWriterCloser) Close() error {
	return nil
}

func (r *libpodRuntime) translateStreams(stdin io.Reader, stdout, stderr io.Writer) *define.AttachStreams {
	return &define.AttachStreams{
		OutputStream: &fakeWriterCloser{stdout},
		ErrorStream:  &fakeWriterCloser{stderr},
		InputStream:  bufio.NewReader(stdin),
		AttachOutput: stdout != nil,
		AttachError:  stderr != nil,
		AttachInput:  stdin != nil,
	}
}

func (r *libpodRuntime) doHookAction(logger log.Interface, ctr *libpod.Container, hook *aranyagopb.ActionMethod) error {
	_ = logger
	switch action := hook.Action.(type) {
	case *aranyagopb.ActionMethod_Exec:
		if cmd := action.Exec.Command; len(cmd) > 0 {
			err := r.execInContainer(ctr, nil, os.Stdout, os.Stderr, nil, cmd, false)
			if err != nil {
				return fmt.Errorf("failed to execute exec hook, code %d: %s", err.Code, err.Description)
			}
		}
	case *aranyagopb.ActionMethod_Http:
	case *aranyagopb.ActionMethod_Socket:
	}
	return nil
}

func (r *libpodRuntime) translateRestartPolicy(policy aranyagopb.RestartPolicy) string {
	switch policy {
	case aranyagopb.RESTART_ALWAYS:
		return libpod.RestartPolicyAlways
	case aranyagopb.RESTART_ON_FAILURE:
		return libpod.RestartPolicyOnFailure
	case aranyagopb.RESTART_NEVER:
		return libpod.RestartPolicyNo
	}

	return libpod.RestartPolicyAlways
}

func (r *libpodRuntime) translatePodStatus(
	podIP string,
	pauseCtr *libpod.Container,
	containers []*libpod.Container,
) (*aranyagopb.PodStatusMsg, error) {
	podUID := pauseCtr.Labels()[constant.ContainerLabelPodUID]
	ctrStatus := make(map[string]*aranyagopb.ContainerStatus)

	for _, ctr := range containers {
		labels := ctr.Labels()
		ctrPodUID := labels[constant.ContainerLabelPodUID]
		name := labels[constant.ContainerLabelPodContainer]
		if name == "" || ctrPodUID != podUID {
			// invalid container, skip
			continue
		}

		status, err := r.translateContainerStatus(ctr)
		if err != nil {
			return nil, err
		}

		ctrStatus[name] = status
	}

	return aranyagopb.NewPodStatusMsg(podUID, podIP, ctrStatus), nil
}

func (r *libpodRuntime) translateContainerStatus(ctr *libpod.Container) (*aranyagopb.ContainerStatus, error) {
	info, err := ctr.Inspect(false)
	if err != nil {
		return nil, err
	}

	return &aranyagopb.ContainerStatus{
		ContainerId:  info.ID,
		ImageId:      info.Image,
		CreatedAt:    info.Created.Format(aranyagoconst.TimeLayout),
		StartedAt:    info.State.StartedAt.Format(aranyagoconst.TimeLayout),
		FinishedAt:   info.State.FinishedAt.Format(aranyagoconst.TimeLayout),
		ExitCode:     info.State.ExitCode,
		RestartCount: info.RestartCount,
		Message:      info.State.Error,
		Reason: func() string {
			switch {
			case info.State.Restarting:
				return "Restarting"
			case info.State.Dead:
				return "Dead"
			case info.State.OOMKilled:
				return "OutOfMemoryKilled"
			case info.State.Running:
				return "Running"
			case info.State.Paused:
				return "Paused"
			}
			return ""
		}(),
	}, nil
}
