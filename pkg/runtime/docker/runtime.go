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

package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/log"
	"arhat.dev/pkg/wellknownerrors"
	dockertype "github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	dockerfilter "github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	dockercopy "github.com/docker/docker/pkg/stdcopy"

	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/arhat/pkg/constant"
	"arhat.dev/arhat/pkg/network"
	"arhat.dev/arhat/pkg/runtime/runtimeutil"
	"arhat.dev/arhat/pkg/types"
	"arhat.dev/arhat/pkg/util/errconv"
)

func NewDockerRuntime(
	ctx context.Context,
	storage types.Storage,
	config *conf.ArhatRuntimeConfig,
) (types.Runtime, error) {
	dialCtxFunc := func(timeout time.Duration) func(
		ctx context.Context, network, addr string,
	) (conn net.Conn, e error) {
		return func(ctx context.Context, network, addr string) (conn net.Conn, e error) {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			var dialer net.Dialer
			if filepath.IsAbs(addr) {
				network = "unix"
				idx := strings.LastIndexByte(addr, ':')
				if idx != -1 {
					addr = addr[:idx]
				}
			}
			return dialer.DialContext(ctx, network, addr)
		}
	}

	runtimeClient, err := dockerclient.NewClientWithOpts(
		dockerclient.WithAPIVersionNegotiation(),
		dockerclient.WithHost(config.EndPoints.Container.Endpoint),
		dockerclient.WithDialContext(dialCtxFunc(config.EndPoints.Container.DialTimeout)),
		dockerclient.FromEnv,
	)
	if err != nil {
		return nil, err
	}

	imageClient := runtimeClient
	if config.EndPoints.Image.Endpoint != config.EndPoints.Container.Endpoint {
		imageClient, err = dockerclient.NewClientWithOpts(
			dockerclient.WithHost(config.EndPoints.Container.Endpoint),
			dockerclient.WithDialContext(dialCtxFunc(config.EndPoints.Image.DialTimeout)),
			dockerclient.FromEnv,
		)
		if err != nil {
			return nil, err
		}
	}

	infoCtx, cancelInfo := context.WithTimeout(ctx, config.EndPoints.Container.ActionTimeout)
	defer cancelInfo()

	versions, err := runtimeClient.ServerVersion(infoCtx)
	if err != nil {
		return nil, err
	}

	version := ""
	for _, ver := range versions.Components {
		if strings.ToLower(ver.Name) == "engine" {
			version = ver.Version
		}
	}

	rt := &dockerRuntime{
		BaseRuntime: runtimeutil.NewBaseRuntime(
			ctx, config, "docker",
			version,
			versions.Os,
			versions.Arch,
			versions.KernelVersion,
		),
		imageClient:   imageClient,
		runtimeClient: runtimeClient,
		storage:       storage,
	}

	rt.networkClient = network.NewNetworkClient(func(subCmd []string, output io.Writer) error {
		logger := rt.Log().WithName("network")
		abbotCtrInfo, err := rt.findAbbotContainer()
		if err != nil {
			if errors.Is(err, wellknownerrors.ErrNotFound) {
				return errors.New("abbot container not found")
			}

			return fmt.Errorf("unable to find abbot container: %w", err)
		}

		cmd := append(strings.Split(abbotCtrInfo.Command, " "), subCmd...)
		msgErr := rt.execInContainer(logger, abbotCtrInfo.ID, nil, output, output, nil, cmd, false)
		if msgErr != nil {
			return fmt.Errorf("unable to execute network command: %v", msgErr)
		}
		return nil
	})

	return rt, nil
}

type dockerRuntime struct {
	*runtimeutil.BaseRuntime

	runtimeClient dockerclient.ContainerAPIClient
	imageClient   dockerclient.ImageAPIClient
	networkClient *network.Client
	storage       types.Storage
}

func (r *dockerRuntime) InitRuntime() error {
	logger := r.Log().WithFields(log.String("action", "init"))
	ctx, cancelInit := r.ActionContext()
	defer cancelInit()

	logger.D("looking up abbot container")
	abbotCtrInfo, err := r.findAbbotContainer()
	if err == nil {
		podUID := abbotCtrInfo.Labels[constant.ContainerLabelPodUID]
		logger.D("looking up pause container for abbot container")
		pauseCtrInfo, err := r.findContainer(podUID, constant.ContainerNamePause)
		if err == nil {
			logger.D("starting pause container for abbot container")
			plainErr := r.runtimeClient.ContainerStart(ctx, pauseCtrInfo.ID, dockertype.ContainerStartOptions{})
			if plainErr != nil {
				logger.I("failed to start pause container for abbot container", log.Error(plainErr))
			}
		}

		logger.D("starting abbot container")
		plainErr := r.runtimeClient.ContainerStart(ctx, abbotCtrInfo.ID, dockertype.ContainerStartOptions{})
		if plainErr != nil {
			return fmt.Errorf("failed to start abbot container: %v", plainErr)
		}
	}

	containers, plainErr := r.runtimeClient.ContainerList(ctx, dockertype.ContainerListOptions{
		Quiet: true,
		All:   true,
	})
	if plainErr != nil {
		return plainErr
	}

	var (
		pauseContainers []dockertype.Container
		workContainers  []dockertype.Container
	)
	for i, ctr := range containers {
		if _, ok := ctr.Labels[constant.ContainerLabelPodUID]; ok {
			switch ctr.Labels[constant.ContainerLabelPodContainerRole] {
			case constant.ContainerRoleInfra:
				pauseContainers = append(pauseContainers, containers[i])
			case constant.ContainerRoleWork:
				workContainers = append(workContainers, containers[i])
			}
		}
	}

	for _, ctr := range pauseContainers {
		logger.D("starting pause container", log.Strings("names", ctr.Names))
		err := r.runtimeClient.ContainerStart(ctx, ctr.ID, dockertype.ContainerStartOptions{})
		if err != nil {
			logger.I("failed to start pause container", log.Strings("names", ctr.Names), log.Error(err))
			return err
		}

		if runtimeutil.IsHostNetwork(ctr.Labels) {
			continue
		}

		pauseCtr, err := r.runtimeClient.ContainerInspect(ctx, ctr.ID)
		if err != nil {
			logger.I("failed to inspect pause container", log.Strings("names", ctr.Names), log.Error(err))
			return err
		}

		err = r.networkClient.RestoreLink(pauseCtr.ID, uint32(pauseCtr.State.Pid))
		if err != nil {
			logger.I("failed to restore container network")
			return err
		}
	}

	for _, ctr := range workContainers {
		logger.D("starting work container", log.Strings("names", ctr.Names))
		err := r.runtimeClient.ContainerStart(ctx, ctr.ID, dockertype.ContainerStartOptions{})
		if err != nil {
			logger.I("failed to start work container", log.Strings("names", ctr.Names), log.Error(err))
		}
	}

	return nil
}

func (r *dockerRuntime) EnsureImages(options *aranyagopb.ImageEnsureOptions) ([]*aranyagopb.Image, error) {
	logger := r.Log().WithFields(log.String("action", "ensureImages"), log.Any("options", options))
	logger.D("ensuring pod container image(s)")

	allImages := map[string]*aranyagopb.ImagePull{
		r.PauseImage: {PullPolicy: aranyagopb.IMAGE_PULL_IF_NOT_PRESENT},
	}

	for imageName, opt := range options.ImagePull {
		allImages[imageName] = opt
	}

	pulledImages, err := r.ensureImages(allImages)
	if err != nil {
		logger.I("failed to ensure container images", log.Error(err))
		return nil, err
	}

	var images []*aranyagopb.Image
	for _, img := range pulledImages {
		var sha256Hash string
		for _, digest := range img.RepoDigests {
			idx := strings.LastIndex(digest, "sha256:")
			if idx > -1 {
				sha256Hash = digest[idx+7:]
			}
		}

		if sha256Hash == "" {
			continue
		}

		images = append(images, &aranyagopb.Image{
			Sha256: sha256Hash,
			Names:  img.RepoTags,
		})
	}

	return images, nil
}

func (r *dockerRuntime) CreateInitContainers(options *aranyagopb.CreateOptions) (*aranyagopb.PodStatus, error) {
	logger := r.Log().WithFields(
		log.String("action", "createInitContainers"),
		log.String("namespace", options.Namespace),
		log.String("name", options.Name),
		log.String("uid", options.PodUid))
	logger.D("creating init containers")

	ctx, cancelCreate := r.RuntimeActionContext()
	defer cancelCreate()

	pauseCtr, podIP, err := r.createPauseContainer(ctx, options)
	if err != nil {
		logger.I("failed to create pause container", log.Error(err))
		return nil, err
	}

	defer func() {
		if err != nil {
			logger.D("deleting pause container due to error")
			err2 := r.deleteContainer(pauseCtr.ID, true)
			if err2 != nil {
				logger.I("failed to delete pause container", log.Error(err2))
			}

			logger.D("cleaning up pod data")
			err2 = runtimeutil.CleanupPodData(
				r.PodDir(options.PodUid),
				r.PodRemoteVolumeDir(options.PodUid, ""),
				r.PodTmpfsVolumeDir(options.PodUid, ""),
				r.storage,
			)
			if err2 != nil {
				logger.E("failed to cleanup pod data", log.Error(err2))
			}
		}
	}()

	// create and wait for init containers
	containers := make(map[string]*aranyagopb.ActionMethod)
	for _, spec := range options.Containers {
		var ctrID string
		ctrID, err = r.createContainer(ctx, options, spec, runtimeutil.SharedNamespaces(pauseCtr.ID, options))
		if err != nil {
			logger.I("failed to create container", log.String("container", spec.Name), log.Error(err))
			return nil, err
		}
		containers[ctrID] = spec.HookPostStart

		defer func(ctrID string) {
			if err != nil {
				logger.D("deleting init container due to error")
				err := r.deleteContainer(ctrID, false)
				if err != nil {
					logger.I("failed to delete init container", log.Error(err))
				}
			}
		}(ctrID)
	}

	wg := new(sync.WaitGroup)
	respCh := make(chan dockertype.ContainerJSON, len(containers))
	errCh := make(chan error, len(containers))
	for ctrID, postStartHook := range containers {
		waitRespCh, waitErrCh := r.runtimeClient.ContainerWait(ctx, ctrID, dockercontainer.WaitConditionNextExit)

		wg.Add(1)
		go func(ctrID string) {
			defer func() {
				wg.Done()

				logger.D("deleting init container", log.String("id", ctrID))
				err := r.deleteContainer(ctrID, false)
				if err != nil {
					logger.I("failed to delete init container", log.String("id", ctrID), log.Error(err))
				}
			}()

			select {
			case resp := <-waitRespCh:
				if resp.StatusCode != 0 {
					if resp.Error != nil {
						errCh <- errors.New(resp.Error.Message)
					} else {
						errCh <- fmt.Errorf("container exited with code %d", resp.StatusCode)
					}
					return
				}

				// init container finished successfully, inspect container info
				ctrInfo, plainErr := r.runtimeClient.ContainerInspect(ctx, ctrID)
				if plainErr != nil {
					errCh <- fmt.Errorf("failed to inspect init container [%s]: %v", ctrID, plainErr)
					return
				}

				respCh <- ctrInfo
			case err := <-waitErrCh:
				errCh <- err
				return
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			}
		}(ctrID)

		logger.D("starting init container")
		err := r.runtimeClient.ContainerStart(ctx, ctrID, dockertype.ContainerStartOptions{})
		if err != nil {
			logger.I("failed to start init container", log.String("id", ctrID), log.Error(err))
			return nil, err
		}

		if postStartHook != nil {
			logger.D("executing post-start hook")
			if err := r.doHookActions(logger, ctrID, postStartHook); err != nil {
				logger.I("failed to execute post-start hook", log.StringError(err.Description))
			}
		}
	}

	// wait for init container operations
	wg.Wait()

	close(respCh)
	close(errCh)

	// check wait operation error
	if err := runtimeutil.CollectErrors(errCh); err != nil {
		return nil, err
	}

	// all init container finished successfully
	var allCtrInfo []*dockertype.ContainerJSON
	for ctrInfo := range respCh {
		info := ctrInfo
		allCtrInfo = append(allCtrInfo, &info)
	}

	return r.translatePodStatus(podIP, pauseCtr, allCtrInfo), nil
}

func (r *dockerRuntime) CreateContainers(options *aranyagopb.CreateOptions) (_ *aranyagopb.PodStatus, err error) {
	logger := r.Log().WithFields(
		log.String("action", "create"),
		log.String("namespace", options.Namespace),
		log.String("name", options.Name),
		log.String("uid", options.PodUid),
	)
	logger.D("creating pod containers")

	ctx, cancelCreate := r.RuntimeActionContext()
	defer func() {
		cancelCreate()

		if err != nil && !errors.Is(err, wellknownerrors.ErrAlreadyExists) {
			logger.D("cleaning up pod data")
			err2 := runtimeutil.CleanupPodData(
				r.PodDir(options.PodUid),
				r.PodRemoteVolumeDir(options.PodUid, ""),
				r.PodTmpfsVolumeDir(options.PodUid, ""),
				r.storage,
			)
			if err2 != nil {
				logger.E("failed to cleanup pod data", log.Error(err2))
			}
		}
	}()

	var (
		pauseCtrInfo *dockertype.ContainerJSON
		podIP        string
	)

	pauseCtr, err := r.findContainer(options.PodUid, constant.ContainerNamePause)
	if err != nil {
		if errors.Is(err, wellknownerrors.ErrNotFound) {
			// need to create pause container
			pauseCtrInfo, podIP, err = r.createPauseContainer(ctx, options)
			if err != nil {
				logger.I("failed to create pause container", log.Error(err))
				return nil, err
			}

			defer func() {
				if err != nil {
					logger.D("deleting pause container due to error")
					err2 := r.deleteContainer(pauseCtrInfo.ID, true)
					if err2 != nil {
						logger.I("failed to delete pause container", log.Error(err2))
					}
				}
			}()
		} else {
			return nil, err
		}
	} else {
		oldPauseCtrInfo, err := r.runtimeClient.ContainerInspect(ctx, pauseCtr.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to inspect existing pause container: %w", err)
		}

		pauseCtrInfo = &oldPauseCtrInfo
	}

	containers := make(map[string]*aranyagopb.ActionMethod)
	for _, spec := range options.Containers {
		ctrLogger := logger.WithFields(log.String("container", spec.Name))
		ctrLogger.D("creating container")
		ctrID, err := r.createContainer(ctx, options, spec, runtimeutil.SharedNamespaces(pauseCtrInfo.ID, options))
		if err != nil {
			ctrLogger.I("failed to create container", log.Error(err))
			return nil, err
		}
		containers[ctrID] = spec.HookPostStart

		defer func(ctrID string) {
			if err != nil {
				ctrLogger.I("delete container due to error")
				if err := r.deleteContainer(ctrID, false); err != nil {
					ctrLogger.I("failed to delete container after start failure", log.Error(err))
				}
			}
		}(ctrID)
	}

	for ctrID, postStartHook := range containers {
		logger.D("starting container", log.String("id", ctrID))
		err := r.runtimeClient.ContainerStart(ctx, ctrID, dockertype.ContainerStartOptions{})
		if err != nil {
			logger.I("failed to start container", log.String("id", ctrID), log.Error(err))
			return nil, err
		}

		if postStartHook != nil {
			logger.D("executing post-start hook")
			if err := r.doHookActions(logger, ctrID, postStartHook); err != nil {
				logger.I("failed to execute post-start hook", log.StringError(err.Description))
			}
		}
	}

	var allCtrInfo []*dockertype.ContainerJSON
	for ctrID := range containers {
		ctrInfo, err := r.runtimeClient.ContainerInspect(ctx, ctrID)
		if err != nil {
			logger.I("failed to inspect docker container", log.Error(err))
			return nil, err
		}
		allCtrInfo = append(allCtrInfo, &ctrInfo)
	}

	return r.translatePodStatus(podIP, pauseCtrInfo, allCtrInfo), nil
}

func (r *dockerRuntime) DeleteContainers(podUID string, containers []string) (*aranyagopb.PodStatus, error) {
	// TODO: implement
	return nil, wellknownerrors.ErrNotSupported
}

func (r *dockerRuntime) DeletePod(options *aranyagopb.DeleteOptions) (_ *aranyagopb.PodStatus, err error) {
	logger := r.Log().WithFields(log.String("action", "delete"), log.Any("options", options))
	logger.D("deleting pod containers")

	ctx, cancelDelete := r.RuntimeActionContext()
	defer func() {
		cancelDelete()

		logger.D("cleaning up pod data")
		err2 := runtimeutil.CleanupPodData(
			r.PodDir(options.PodUid),
			r.PodRemoteVolumeDir(options.PodUid, ""),
			r.PodTmpfsVolumeDir(options.PodUid, ""),
			r.storage,
		)
		if err2 != nil {
			logger.E("failed to cleanup pod data", log.Error(err2))
		}
	}()

	containers, err := r.runtimeClient.ContainerList(ctx, dockertype.ContainerListOptions{
		Quiet: true,
		All:   true,
		Filters: dockerfilter.NewArgs(
			dockerfilter.Arg("label", fmt.Sprintf("%s=%s", constant.ContainerLabelPodUID, options.PodUid)),
		),
	})
	if err != nil && !dockerclient.IsErrNotFound(err) {
		logger.I("failed to list containers", log.Error(err))
		return nil, err
	}

	// delete work containers first
	pauseCtrIndex := -1
	for i, ctr := range containers {
		// find pause container
		if ctr.Labels[constant.ContainerLabelPodContainerRole] == constant.ContainerRoleInfra {
			pauseCtrIndex = i
			break
		}
	}
	lastIndex := len(containers) - 1
	// swap pause container to last
	if pauseCtrIndex != -1 {
		containers[lastIndex], containers[pauseCtrIndex] = containers[pauseCtrIndex], containers[lastIndex]
	}

	for i, ctr := range containers {
		logger.D("deleting container", log.Strings("names", ctr.Names))

		isPauseCtr := false
		if i == len(containers)-1 {
			// last one, is deleting pause container, we need to delete network first
			isPauseCtr = true
		}

		name := ctr.Labels[constant.ContainerLabelPodContainer]
		if options.HookPreStop != nil {
			if hook, ok := options.HookPreStop[name]; ok {
				logger.D("executing pre-stop hook", log.String("name", name))
				if err := r.doHookActions(logger, ctr.ID, hook); err != nil {
					logger.I("failed to execute pre-stop hook", log.StringError(err.Description))
				}
			}
		}

		err = r.deleteContainer(ctr.ID, isPauseCtr)
		if err != nil {
			return nil, fmt.Errorf("failed to delete container: %w", err)
		}
	}

	logger.D("pod deleted")
	return aranyagopb.NewPodStatus(options.PodUid, "", nil), nil
}

func (r *dockerRuntime) ListPods(options *aranyagopb.ListOptions) ([]*aranyagopb.PodStatus, error) {
	logger := r.Log().WithFields(log.String("action", "list"), log.Any("options", options))
	logger.D("listing pods")

	ctx, cancelList := r.RuntimeActionContext()
	defer cancelList()

	filter := dockerfilter.NewArgs()
	if !options.All {
		if options.Namespace != "" {
			filter.Add("label", constant.ContainerLabelPodNamespace+"="+options.Namespace)
		}

		if options.Name != "" {
			filter.Add("label", constant.ContainerLabelPodName+"="+options.Name)
		}
	}

	containers, err := r.runtimeClient.ContainerList(ctx, dockertype.ContainerListOptions{
		All:     options.All,
		Quiet:   true,
		Filters: filter,
	})
	if err != nil {
		logger.I("failed to list containers", log.Error(err))
		return nil, err
	}

	var (
		results []*aranyagopb.PodStatus
		// podUID -> pause container
		pauseContainers = make(map[string]dockertype.Container)
		// podUID -> containers
		podContainers = make(map[string][]dockertype.Container)
	)

	for _, ctr := range containers {
		podUID, hasUID := ctr.Labels[constant.ContainerLabelPodUID]
		if !hasUID {
			// not the container created by us
			continue
		}

		role, hasRole := ctr.Labels[constant.ContainerLabelPodContainerRole]
		if hasRole && role == constant.ContainerRoleInfra {
			pauseContainers[podUID] = ctr
			continue
		}

		podContainers[podUID] = append(podContainers[podUID], ctr)
	}

	// one pause container represents on Pod
	for podUID, pauseContainer := range pauseContainers {
		logger.D("inspecting pause container")
		pauseCtrSpec, err := r.runtimeClient.ContainerInspect(ctx, pauseContainer.ID)
		if err != nil {
			logger.I("failed to inspect pause container", log.Error(err))
			return nil, err
		}

		var containersInfo []*dockertype.ContainerJSON
		for _, ctr := range podContainers[podUID] {
			logger.D("inspecting work container")
			var ctrInfo dockertype.ContainerJSON
			ctrInfo, err = r.runtimeClient.ContainerInspect(ctx, ctr.ID)
			if err != nil {
				logger.I("failed to inspect work container", log.Error(err))
				return nil, err
			}
			containersInfo = append(containersInfo, &ctrInfo)
		}

		var podIP string
		if !runtimeutil.IsHostNetwork(pauseCtrSpec.Config.Labels) {
			podIP, err = r.networkClient.GetAddress(uint32(pauseCtrSpec.State.Pid))
			if err != nil {
				return nil, err
			}
		}
		results = append(results, r.translatePodStatus(podIP, &pauseCtrSpec, containersInfo))
	}

	return results, nil
}

func (r *dockerRuntime) ExecInContainer(
	podUID, container string,
	stdin io.Reader,
	stdout, stderr io.Writer,
	resizeCh <-chan *aranyagopb.TtyResizeOptions,
	command []string, tty bool,
) *aranyagopb.Error {
	logger := r.Log().WithFields(
		log.String("uid", podUID),
		log.String("container", container),
		log.String("action", "exec"),
	)
	logger.D("exec in pod container")

	ctr, err := r.findContainer(podUID, container)
	if err != nil {
		return errconv.ToConnectivityError(err)
	}

	return r.execInContainer(logger, ctr.ID, stdin, stdout, stderr, resizeCh, command, tty)
}

func (r *dockerRuntime) AttachContainer(
	podUID, container string,
	stdin io.Reader,
	stdout, stderr io.Writer,
	resizeCh <-chan *aranyagopb.TtyResizeOptions,
) error {
	logger := r.Log().WithFields(
		log.String("action", "attach"),
		log.String("uid", podUID),
		log.String("container", container),
	)
	logger.D("attach to pod container")

	ctr, err := r.findContainer(podUID, container)
	if err != nil {
		return err
	}

	ctx, cancelAttach := r.ActionContext()
	defer cancelAttach()

	resp, err := r.runtimeClient.ContainerAttach(ctx, ctr.ID, dockertype.ContainerAttachOptions{
		Stream: true,
		Stdin:  stdin != nil,
		Stdout: stdout != nil,
		Stderr: stderr != nil,
	})
	if err != nil {
		return fmt.Errorf("failed to attach container: %w", err)
	}
	defer func() { _ = resp.Conn.Close() }()

	if stdin != nil {
		go func() {
			_, err2 := io.Copy(resp.Conn, stdin)
			if err2 != nil {
				logger.I("exception happened in write routine", log.Error(err2))
			}
		}()
	}

	go func() {
		for {
			select {
			case size, more := <-resizeCh:
				if !more {
					return
				}

				err2 := r.runtimeClient.ContainerResize(ctx, ctr.ID, dockertype.ResizeOptions{
					Height: uint(size.Rows),
					Width:  uint(size.Cols),
				})
				if err2 != nil {
					logger.I("exception happened in tty resize routine", log.Error(err2))
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	if stderr != nil {
		_, err = dockercopy.StdCopy(stdout, stderr, resp.Reader)
	} else {
		_, err = io.Copy(stdout, resp.Reader)
	}
	if err != nil {
		logger.I("exception happened in read routine", log.Error(err))
	}

	return nil
}

func (r *dockerRuntime) GetContainerLogs(
	podUID string,
	options *aranyagopb.LogOptions,
	stdout, stderr io.WriteCloser,
	logCtx context.Context,
) error {
	logger := r.Log().WithFields(
		log.String("action", "log"),
		log.String("uid", podUID),
		log.Any("options", options),
	)
	logger.D("fetching pod container logs")

	ctr, err := r.findContainer(podUID, options.Container)
	if err != nil {
		return err
	}

	var since, tail string
	if options.Since != "" {
		since = options.Since
	}

	if options.TailLines > 0 {
		tail = strconv.FormatInt(options.TailLines, 10)
	}

	logReader, err := r.runtimeClient.ContainerLogs(logCtx, ctr.ID, dockertype.ContainerLogsOptions{
		ShowStdout: stdout != nil,
		ShowStderr: stderr != nil,
		Since:      since,
		Timestamps: options.Timestamp,
		Follow:     options.Follow,
		Tail:       tail,
		Details:    false,
	})
	if err != nil {
		return fmt.Errorf("failed to read container logs: %w", err)
	}
	defer func() { _ = logReader.Close() }()

	_, err = dockercopy.StdCopy(stdout, stderr, logReader)
	if err != nil {
		return fmt.Errorf("exception happened when reading logs: %w", err)
	}

	return nil
}

func (r *dockerRuntime) PortForward(podUID string, protocol string, port int32, downstream io.ReadWriter) error {
	logger := r.Log().WithFields(
		log.String("action", "portforward"),
		log.String("proto", protocol),
		log.Int32("port", port),
		log.String("uid", podUID),
	)

	logger.D("port-forwarding to pod container")
	pauseCtr, err := r.findContainer(podUID, constant.ContainerNamePause)
	if err != nil {
		return err
	}

	// TODO: fix address discovery
	if pauseCtr.NetworkSettings == nil {
		return fmt.Errorf("failed to find network settings: %w", err)
	}

	address := ""
	for name, endpoint := range pauseCtr.NetworkSettings.Networks {
		if name == "bridge" {
			address = endpoint.IPAddress
			break
		}
	}

	if address == "" {
		return fmt.Errorf("failed to find container bridge address: %w", err)
	}

	ctx, cancel := r.ActionContext()
	defer cancel()

	return runtimeutil.PortForward(ctx, address, protocol, port, downstream)
}

func (r *dockerRuntime) UpdateContainerNetwork(options *aranyagopb.NetworkOptions) ([]*aranyagopb.PodStatus, error) {
	logger := r.Log().WithFields(log.String("action", "updateContainerNetwork"))

	ctx, cancelUpdate := r.ActionContext()
	defer cancelUpdate()

	logger.D("looking up abbot container")
	_, err := r.findAbbotContainer()
	if err != nil {
		if errors.Is(err, wellknownerrors.ErrNotFound) {
			return nil, fmt.Errorf(
				"cluster network relies on abbot container but not found: %w",
				wellknownerrors.ErrNotSupported,
			)
		}

		return nil, fmt.Errorf("failed to find abbot container: %w", err)
	}

	logger.D("looking up all pause containers")
	pauseCtrs, err := r.runtimeClient.ContainerList(ctx, dockertype.ContainerListOptions{
		Quiet: true,
		Filters: dockerfilter.NewArgs(
			dockerfilter.Arg("label",
				fmt.Sprintf("%s=%s",
					constant.ContainerLabelPodContainerRole,
					constant.ContainerRoleInfra,
				),
			),
		),
	})
	if err != nil {
		logger.I("failed to list all pause containers", log.Error(err))
		return nil, err
	}

	var result []*aranyagopb.PodStatus
	for _, ctr := range pauseCtrs {
		podUID, ok := ctr.Labels[constant.ContainerLabelPodUID]
		if !ok {
			continue
		}

		if ctr.HostConfig.NetworkMode == "" {
			logger.D("inspecting pause container with cluster network")
			ctrSpec, err := r.runtimeClient.ContainerInspect(ctx, ctr.ID)
			if err != nil {
				logger.I("failed to inspect container")
				return nil, err
			}

			logger.D("ensuring pause container network", log.String("name", ctrSpec.Name))
			ip, err := r.networkClient.EnsureAddress(ctrSpec.ID, uint32(ctrSpec.State.Pid), options)
			if err != nil {
				return nil, err
			}

			result = append(result, aranyagopb.NewPodStatus(podUID, ip, nil))
		}
	}
	return result, nil
}
