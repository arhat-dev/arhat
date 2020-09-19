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

package docker

import (
	"bytes"
	"time"

	"arhat.dev/aranya-proto/aranyagopb/aranyagoconst"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/log"
	dockertype "github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"

	"arhat.dev/arhat/pkg/constant"
)

var (
	restartAlways    = dockercontainer.RestartPolicy{Name: "always"}
	restartOnFailure = dockercontainer.RestartPolicy{Name: "on-failure"}
	restartNever     = dockercontainer.RestartPolicy{Name: "no"}
)

func (r *dockerRuntime) translateRestartPolicy(policy aranyagopb.RestartPolicy) dockercontainer.RestartPolicy {
	switch policy {
	case aranyagopb.RESTART_ALWAYS:
		return restartAlways
	case aranyagopb.RESTART_ON_FAILURE:
		return restartOnFailure
	case aranyagopb.RESTART_NEVER:
		return restartNever
	}

	return restartAlways
}

func (r *dockerRuntime) translatePodStatus(
	podIP string,
	pauseContainer *dockertype.ContainerJSON,
	containers []*dockertype.ContainerJSON,
) *aranyagopb.PodStatusMsg {
	podUID := pauseContainer.Config.Labels[constant.ContainerLabelPodUID]
	ctrStatus := make(map[string]*aranyagopb.ContainerStatus)

	for _, ctr := range containers {
		ctrPodUID := ctr.Config.Labels[constant.ContainerLabelPodUID]
		name := ctr.Config.Labels[constant.ContainerLabelPodContainer]
		if name == "" || ctrPodUID != podUID {
			// invalid container, skip
			continue
		}

		status := r.translateContainerStatus(ctr)
		ctrStatus[name] = status
	}

	return aranyagopb.NewPodStatusMsg(podUID, podIP, ctrStatus)
}

func (r *dockerRuntime) translateContainerStatus(
	ctrInfo *dockertype.ContainerJSON,
) *aranyagopb.ContainerStatus {
	ctrCreatedAt, _ := time.Parse(time.RFC3339Nano, ctrInfo.Created)
	ctrStartedAt, _ := time.Parse(time.RFC3339Nano, ctrInfo.State.StartedAt)
	ctrFinishedAt, _ := time.Parse(time.RFC3339Nano, ctrInfo.State.FinishedAt)

	return &aranyagopb.ContainerStatus{
		ContainerId: r.Name() + "://" + ctrInfo.ID,
		ImageId:     ctrInfo.Image,
		CreatedAt:   ctrCreatedAt.Format(aranyagoconst.TimeLayout),
		StartedAt:   ctrStartedAt.Format(aranyagoconst.TimeLayout),
		FinishedAt:  ctrFinishedAt.Format(aranyagoconst.TimeLayout),
		ExitCode: func() int32 {
			if ctrInfo.State != nil {
				return int32(ctrInfo.State.ExitCode)
			}
			return 0
		}(),
		RestartCount: int32(ctrInfo.RestartCount),
	}
}

func (r *dockerRuntime) doHookActions(
	logger log.Interface,
	ctrID string,
	hook *aranyagopb.ActionMethod,
) *aranyagopb.ErrorMsg {
	switch action := hook.Action.(type) {
	case *aranyagopb.ActionMethod_Exec:
		if cmd := action.Exec.Command; len(cmd) > 0 {
			buf := new(bytes.Buffer)
			err := r.execInContainer(logger, ctrID, nil, buf, buf, nil, cmd, false)
			if err != nil {
				return err
			}
		}
	case *aranyagopb.ActionMethod_Http:
	case *aranyagopb.ActionMethod_Socket:
	}

	return nil
}
