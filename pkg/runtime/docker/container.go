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
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/log"
	"arhat.dev/pkg/wellknownerrors"
	dockertype "github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	dockerfilter "github.com/docker/docker/api/types/filters"
	dockermount "github.com/docker/docker/api/types/mount"
	dockerstrslice "github.com/docker/docker/api/types/strslice"
	dockerclient "github.com/docker/docker/client"
	dockercopy "github.com/docker/docker/pkg/stdcopy"
	dockernat "github.com/docker/go-connections/nat"

	"arhat.dev/arhat/pkg/constant"
	"arhat.dev/arhat/pkg/runtime/runtimeutil"
	"arhat.dev/arhat/pkg/types"
)

func (r *dockerRuntime) listContainersByLabels(labels map[string]string) ([]*dockertype.Container, error) {
	findCtx, cancelFind := r.RuntimeActionContext()
	defer cancelFind()

	filters := dockerfilter.NewArgs()
	for k, v := range labels {
		filters.Add("label", fmt.Sprintf("%s=%s", k, v))
	}

	containers, err := r.runtimeClient.ContainerList(findCtx, dockertype.ContainerListOptions{
		Quiet:   true,
		Filters: filters,
	})
	if err != nil {
		return nil, err
	}

	result := make([]*dockertype.Container, len(containers))
	for i := range containers {
		result[i] = &containers[i]
	}

	return result, nil
}

func (r *dockerRuntime) listPauseContainers() ([]*dockertype.Container, error) {
	return r.listContainersByLabels(
		map[string]string{
			constant.ContainerLabelPodContainerRole: constant.ContainerRoleInfra,
		},
	)
}

func (r *dockerRuntime) findContainerByLabels(labels map[string]string) (*dockertype.Container, error) {
	containers, err := r.listContainersByLabels(labels)
	if err != nil {
		return nil, err
	}

	if len(containers) == 0 {
		return nil, wellknownerrors.ErrNotFound
	}

	return containers[0], nil
}

func (r *dockerRuntime) findContainer(podUID, container string) (*dockertype.Container, error) {
	return r.findContainerByLabels(map[string]string{
		constant.ContainerLabelPodUID:       podUID,
		constant.ContainerLabelPodContainer: container,
	})
}

func (r *dockerRuntime) findAbbotContainer() (*dockertype.Container, error) {
	return r.findContainerByLabels(runtimeutil.AbbotMatchLabels())
}

func (r *dockerRuntime) execInContainer(
	logger log.Interface,
	ctrID string,
	stdin io.Reader,
	stdout, stderr io.Writer,
	resizeCh <-chan *aranyagopb.ContainerTerminalResizeCmd,
	command []string,
	tty bool,
) *aranyagopb.ErrorMsg {
	execCtx, cancelExec := r.ActionContext()
	defer cancelExec()

	var plainErr error
	resp, plainErr := r.runtimeClient.ContainerExecCreate(execCtx, ctrID, dockertype.ExecConfig{
		Tty:          tty,
		AttachStdin:  stdin != nil,
		AttachStdout: stdout != nil,
		AttachStderr: stderr != nil,
		Cmd:          command,
	})
	if plainErr != nil {
		logger.I("failed to exec create", log.Error(plainErr))
		return aranyagopb.NewCommonErrorMsg(plainErr.Error())
	}

	attachResp, plainErr := r.runtimeClient.ContainerExecAttach(execCtx, resp.ID, dockertype.ExecStartCheck{Tty: tty})
	if plainErr != nil {
		logger.I("failed to exec attach", log.Error(plainErr))
		return aranyagopb.NewCommonErrorMsg(plainErr.Error())
	}
	defer func() { _ = attachResp.Conn.Close() }()

	if stdin != nil {
		go func() {
			_, err := io.Copy(attachResp.Conn, stdin)
			if err != nil {
				logger.I("exception happened in write routine", log.Error(err))
			}
		}()
	}

	if tty && resizeCh != nil {
		go func() {
			defer logger.I("finished tty resize routine")

			for {
				select {
				case size, more := <-resizeCh:
					if !more {
						return
					}

					err := r.runtimeClient.ContainerExecResize(execCtx, resp.ID, dockertype.ResizeOptions{
						Height: uint(size.Rows),
						Width:  uint(size.Cols),
					})
					if err != nil {
						// DO NOT break here
						logger.I("exception happened in tty resize routine", log.Error(err))
					}
				case <-execCtx.Done():
					return
				}
			}
		}()
	}

	// Here, we will only wait for the output
	// since input (stdin) and resize (tty) are optional
	// and kubectl doesn't have a detach option, so the stdout will always be there
	// once this function call returned, base_runtime will close everything related

	var stdOut, stdErr io.Writer
	stdOut, stdErr = stdout, stderr
	if stdout == nil {
		stdOut = ioutil.Discard
	}
	if stderr == nil {
		stdErr = ioutil.Discard
	}

	if tty {
		_, plainErr = io.Copy(stdOut, attachResp.Reader)
	} else {
		_, plainErr = dockercopy.StdCopy(stdOut, stdErr, attachResp.Reader)
	}

	if plainErr != nil {
		logger.I("exception happened in read routine", log.Error(plainErr))
	}

	return nil
}

// nolint:goconst
func (r *dockerRuntime) createPauseContainer(
	ctx context.Context,
	options *aranyagopb.PodEnsureCmd,
) (ctrInfo *dockertype.ContainerJSON, podIPv4, podIPv6 string, err error) {
	_, err = r.findContainer(options.PodUid, constant.ContainerNamePause)
	if err == nil {
		return nil, "", "", wellknownerrors.ErrAlreadyExists
	} else if !errors.Is(err, wellknownerrors.ErrNotFound) {
		return nil, "", "", err
	}

	// refuse to create pod using cluster network if no abbot found
	if !options.HostNetwork {
		_, err = r.findAbbotContainer()
		if err != nil {
			return nil, "", "", fmt.Errorf("abbot container required but not found: %w", err)
		}
	}

	var (
		hosts        []string
		exposedPorts = make(dockernat.PortSet)
		portBindings = make(dockernat.PortMap)
		hostname     string
		logger       = r.Log().WithFields(log.String("action", "createPauseContainer"))
	)

	switch {
	case options.HostNetwork:
		hostname = ""
	case options.Hostname != "":
		hostname = options.Hostname
	default:
		hostname = options.Name
	}

	for k, v := range options.Network.Hosts {
		hosts = append(hosts, fmt.Sprintf("%s:%s", k, v))
	}

	if !options.HostNetwork {
		for _, port := range options.Network.Ports {
			var ctrPort dockernat.Port
			ctrPort, err = dockernat.NewPort(port.Protocol, strconv.FormatInt(int64(port.ContainerPort), 10))
			if err != nil {
				return nil, "", "", err
			}

			exposedPorts[ctrPort] = struct{}{}
			if !options.HostNetwork {
				portBindings[ctrPort] = []dockernat.PortBinding{{
					HostPort: strconv.FormatInt(int64(port.HostPort), 10),
				}}
			}
		}
	}

	pauseCtrName := runtimeutil.GetContainerName(options.Namespace, options.Name, constant.ContainerNamePause)
	pauseCtr, err := r.runtimeClient.ContainerCreate(ctx,
		&dockercontainer.Config{
			Image:           r.PauseImage,
			Entrypoint:      dockerstrslice.StrSlice{r.PauseCommand},
			ExposedPorts:    exposedPorts,
			Hostname:        hostname,
			NetworkDisabled: !options.HostNetwork,
			Labels:          runtimeutil.ContainerLabels(options, constant.ContainerNamePause),
		},
		&dockercontainer.HostConfig{
			Resources: dockercontainer.Resources{
				MemorySwap: 0,
				CPUShares:  2,
			},

			ExtraHosts: hosts,
			Mounts:     []dockermount.Mount{},

			PortBindings:  portBindings,
			RestartPolicy: r.translateRestartPolicy(aranyagopb.RESTART_ALWAYS),

			// kernel namespaces
			NetworkMode: func() dockercontainer.NetworkMode {
				if options.HostNetwork {
					return "host"
				}
				return ""
			}(),
			IpcMode: func() dockercontainer.IpcMode {
				if options.HostIpc {
					return "host"
				}
				return "shareable"
			}(),
			PidMode: func() dockercontainer.PidMode {
				if options.HostPid {
					return "host"
				}
				return "container"
			}(),
		},
		nil, pauseCtrName)

	if err != nil {
		return nil, "", "", err
	}
	defer func() {
		if err != nil {
			err2 := r.deleteContainer(pauseCtr.ID, true)
			if err2 != nil {
				logger.I("failed to delete pause container when error happened", log.Error(err2))
			}
		}
	}()

	err = r.runtimeClient.ContainerStart(ctx, pauseCtr.ID, dockertype.ContainerStartOptions{})
	if err != nil {
		return nil, "", "", err
	}

	pauseCtrSpec, err := r.runtimeClient.ContainerInspect(ctx, pauseCtr.ID)
	if err != nil {
		return nil, "", "", err
	}

	// handle cni network setup
	if !options.HostNetwork {
		podIPv4, podIPv6, err = r.networkClient.EnsurePodNetwork(
			options.Namespace, options.Name, pauseCtr.ID, uint32(pauseCtrSpec.State.Pid), options.Network,
		)
		if err != nil {
			return nil, "", "", err
		}
	}

	return &pauseCtrSpec, podIPv4, podIPv6, nil
}

func (r *dockerRuntime) createContainer(
	ctx context.Context,
	options *aranyagopb.PodEnsureCmd,
	spec *aranyagopb.ContainerSpec,
	ns map[string]string,
) (ctrID string, err error) {
	_, err = r.findContainer(options.PodUid, spec.Name)
	if err == nil {
		return "", wellknownerrors.ErrAlreadyExists
	} else if !errors.Is(err, wellknownerrors.ErrNotFound) {
		return "", err
	}

	var (
		userAndGroup      string
		containerVolumes  = make(map[string]struct{})
		containerBinds    []string
		mounts            []dockermount.Mount
		envs              []string
		hostPaths         = options.GetVolumes().GetHostPaths()
		volumeData        = options.GetVolumes().GetVolumeData()
		containerFullName = runtimeutil.GetContainerName(options.Namespace, options.Name, spec.Name)
		healthCheck       *dockercontainer.HealthConfig
		maskedPaths       []string
		readonlyPaths     []string
	)
	// generalize to avoid panic
	if hostPaths == nil {
		hostPaths = make(map[string]string)
	}

	if volumeData == nil {
		volumeData = make(map[string]*aranyagopb.NamedData)
	}

	for k, v := range spec.Envs {
		envs = append(envs, k+"="+v)
	}

	switch spec.GetSecurity().GetProcMountKind() {
	case aranyagopb.PROC_MOUNT_DEFAULT:
		maskedPaths = nil
		readonlyPaths = nil
	case aranyagopb.PROC_MOUNT_UNMASKED:
		maskedPaths = make([]string, 0)
		readonlyPaths = make([]string, 0)
	}

	for volName, volMountSpec := range spec.Mounts {
		var source string

		containerVolumes[volMountSpec.MountPath] = struct{}{}
		// check if it is host volume or emptyDir
		hostPath, isHostVol := hostPaths[volName]
		if isHostVol {
			source, err = runtimeutil.ResolveHostPathMountSource(
				hostPath, options.PodUid, volName,
				volMountSpec.Remote, r.RuntimeConfig,
			)
			if err != nil {
				return "", err
			}

			if volMountSpec.Remote {
				// for remote volume, hostPath is the aranya pod host path
				err = r.storage.Mount(
					hostPath, source, r.handleStorageFailure(options.PodUid),
				)
				if err != nil {
					return "", err
				}
			}
		}

		// check if it is vol data (from configMap, Secret)
		if volData, isVolData := volumeData[volName]; isVolData {
			if dataMap := volData.GetDataMap(); dataMap != nil {
				dir := r.PodBindVolumeDir(options.PodUid, volName)
				if err = os.MkdirAll(dir, 0750); err != nil {
					return "", err
				}

				source, err = volMountSpec.Ensure(dir, dataMap)
				if err != nil {
					return "", err
				}
			}
		}

		mounts = append(mounts, dockermount.Mount{
			Type:        dockermount.TypeBind,
			Source:      source,
			Target:      filepath.Join(volMountSpec.MountPath, volMountSpec.SubPath),
			ReadOnly:    volMountSpec.ReadOnly,
			Consistency: dockermount.ConsistencyFull,
		})
	}

	if netOpts := options.Network; len(netOpts.NameServers) != 0 {
		resolvConfFile := r.PodResolvConfFile(options.PodUid)
		if err = os.MkdirAll(filepath.Dir(resolvConfFile), 0750); err != nil {
			return "", err
		}

		var data []byte
		data, err = r.networkClient.CreateResolvConf(
			netOpts.NameServers, netOpts.SearchDomains, netOpts.DnsOptions,
		)
		if err != nil {
			return "", err
		}

		if err = ioutil.WriteFile(resolvConfFile, data, 0440); err != nil {
			return "", err
		}

		mounts = append(mounts, dockermount.Mount{
			Type:        dockermount.TypeBind,
			Source:      resolvConfFile,
			Target:      "/etc/resolv.conf",
			Consistency: dockermount.ConsistencyFull,
		})
	}

	if spec.Security != nil {
		builder := &strings.Builder{}
		if uid := spec.Security.GetUser(); uid != -1 {
			builder.WriteString(strconv.FormatInt(uid, 10))
		}

		if gid := spec.Security.GetGroup(); gid != -1 {
			builder.WriteString(":")
			builder.WriteString(strconv.FormatInt(gid, 10))
		}
		userAndGroup = builder.String()
	}

	if probe := spec.LivenessCheck; probe != nil && probe.Method != nil {
		switch action := spec.LivenessCheck.Method.Action.(type) {
		case *aranyagopb.ContainerAction_Exec_:
			healthCheck = &dockercontainer.HealthConfig{
				Test:        append([]string{"CMD"}, action.Exec.Command...),
				Interval:    time.Duration(probe.ProbeInterval),
				Timeout:     time.Duration(probe.ProbeTimeout),
				StartPeriod: time.Duration(probe.InitialDelay),
				Retries:     int(probe.FailureThreshold),
				// TODO: implement success threshold
			}
		case *aranyagopb.ContainerAction_Socket_:
			// TODO: implement
		case *aranyagopb.ContainerAction_Http:
			// TODO: implement
		}
	}

	containerConfig := &dockercontainer.Config{
		User: userAndGroup,

		Tty:       spec.Tty,
		OpenStdin: spec.Stdin,
		StdinOnce: spec.StdinOnce,

		Env: envs,

		Entrypoint: spec.Command,
		Cmd:        spec.Args,

		Healthcheck: healthCheck,

		Image:      spec.Image,
		Volumes:    containerVolumes,
		WorkingDir: spec.WorkingDir,

		Labels:     runtimeutil.ContainerLabels(options, spec.Name),
		StopSignal: "SIGTERM",
	}
	hostConfig := &dockercontainer.HostConfig{
		Resources: dockercontainer.Resources{MemorySwap: 0, CPUShares: 2},
		// volume mounts
		Binds:  containerBinds,
		Mounts: mounts,

		RestartPolicy: r.translateRestartPolicy(options.RestartPolicy),
		// share namespaces
		NetworkMode: dockercontainer.NetworkMode(ns["net"]),
		IpcMode:     dockercontainer.IpcMode(ns["ipc"]),
		UTSMode:     dockercontainer.UTSMode(ns["uts"]),
		UsernsMode:  dockercontainer.UsernsMode(ns["user"]),
		PidMode:     dockercontainer.PidMode(ns["pid"]),

		// security options
		Privileged:     spec.Security.GetPrivileged(),
		CapAdd:         spec.Security.GetCapsAdd(),
		CapDrop:        spec.Security.GetCapsDrop(),
		ReadonlyRootfs: spec.Security.GetReadOnlyRootfs(),
		Sysctls:        options.GetSecurity().GetSysctls(),

		MaskedPaths:   maskedPaths,
		ReadonlyPaths: readonlyPaths,
	}

	ctr, err := r.runtimeClient.ContainerCreate(ctx, containerConfig, hostConfig, nil, containerFullName)
	if err != nil {
		return "", err
	}

	return ctr.ID, nil
}

// deleteContainer return nil if container not found or deleted successfully
func (r *dockerRuntime) deleteContainer(containerID string, isPauseCtr bool) error {
	if isPauseCtr {
		// network manager is available
		pauseCtr, err := r.runtimeClient.ContainerInspect(context.Background(), containerID)
		if err != nil {
			if dockerclient.IsErrNotFound(err) {
				// container already deleted, no more effort
				return nil
			}
		}

		if runtimeutil.IsAbbotPod(pauseCtr.Config.Labels) {
			var containers []*dockertype.Container
			containers, err = r.listPauseContainers()
			if err != nil {
				return err
			}

			for _, ctr := range containers {
				if !runtimeutil.IsHostNetwork(ctr.Labels) {
					return wellknownerrors.ErrInvalidOperation
				}
			}
		}

		if !runtimeutil.IsHostNetwork(pauseCtr.Config.Labels) {
			err = r.networkClient.DeletePodNetwork(pauseCtr.ID, uint32(pauseCtr.State.Pid))
			if err != nil {
				return err
			}
		}
	}

	// stop with best effort
	timeout := time.Duration(0)
	_ = r.runtimeClient.ContainerStop(context.Background(), containerID, &timeout)

	err := r.runtimeClient.ContainerRemove(context.Background(), containerID, dockertype.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})

	if err != nil && !dockerclient.IsErrNotFound(err) {
		return err
	}
	return nil
}

func (r *dockerRuntime) handleStorageFailure(podUID string) types.StorageFailureHandleFunc {
	logger := r.Log().WithFields(log.String("module", "storage"), log.String("podUID", podUID))
	return func(remotePath, mountPoint string, err error) {
		if err != nil {
			logger.I("storage mounter exited", log.Error(err))
		}

		_, e := r.findContainer(podUID, constant.ContainerNamePause)
		if errors.Is(e, wellknownerrors.ErrNotFound) {
			logger.D("pod not found, no more remount action")
			return
		}

		err = r.storage.Mount(remotePath, mountPoint, r.handleStorageFailure(podUID))
		if err != nil {
			logger.I("failed to mount remote volume", log.Error(err))
		}
	}
}
