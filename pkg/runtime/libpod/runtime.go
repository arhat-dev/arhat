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
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/log"
	"arhat.dev/pkg/wellknownerrors"
	libpodconfig "github.com/containers/common/pkg/config"
	"github.com/containers/libpod/v2/libpod"
	"github.com/containers/libpod/v2/libpod/define"
	libpodimage "github.com/containers/libpod/v2/libpod/image"
	libpodversion "github.com/containers/libpod/v2/version"
	"k8s.io/client-go/tools/remotecommand"

	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/arhat/pkg/constant"
	"arhat.dev/arhat/pkg/network"
	"arhat.dev/arhat/pkg/runtime/runtimeutil"
	"arhat.dev/arhat/pkg/types"
	"arhat.dev/arhat/pkg/util/errconv"
	"arhat.dev/arhat/pkg/util/sysinfo"
	"arhat.dev/arhat/pkg/version"
)

func NewLibpodRuntime(
	ctx context.Context,
	storage types.Storage,
	config *conf.ArhatRuntimeConfig,
) (types.Runtime, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	opts := []libpod.RuntimeOption{
		libpod.WithNamespace(config.ManagementNamespace),
		libpod.WithDefaultInfraImage(config.PauseImage),
		libpod.WithDefaultInfraCommand(config.PauseCommand),
		libpod.WithDefaultTransport("docker://"),
	}

	var (
		runtimeClient *libpod.Runtime
		err           error
	)

	// path to the libpod.conf
	configFile, ok := os.LookupEnv("ARHAT_LIBPOD_CONFIG")
	if ok && configFile != "" {
		var rtConfig *libpodconfig.Config
		rtConfig, err = libpodconfig.NewConfig(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load custom libpod config: %w", err)
		}
		runtimeClient, err = libpod.NewRuntimeFromConfig(ctx, rtConfig)
	} else {
		// try to create runtime with default options
		runtimeClient, err = libpod.NewRuntime(ctx, opts...)
	}
	if err != nil {
		return nil, err
	}

	imageClient := runtimeClient.ImageRuntime()
	if imageClient == nil {
		return nil, errors.New("empty image client")
	}

	rt := &libpodRuntime{
		BaseRuntime: runtimeutil.NewBaseRuntime(
			ctx, config, "libpod",
			libpodversion.Version,
			runtime.GOOS, version.Arch(), sysinfo.GetKernelVersion(),
		),

		runtimeClient: runtimeClient,
		imageClient:   imageClient,
		storage:       storage,
	}

	rt.networkClient = network.NewNetworkClient(rt.handleAbbotExec)

	return rt, nil
}

type libpodRuntime struct {
	*runtimeutil.BaseRuntime

	runtimeClient *libpod.Runtime
	imageClient   *libpodimage.Runtime
	networkClient *network.Client
	storage       types.Storage
}

func (r *libpodRuntime) handleAbbotExec(subCmd []string, output io.Writer) error {
	logger := r.Log().WithName("network")
	ctr, err := r.findAbbotContainer()
	if err != nil {
		return err
	}

	// try to start abbot container if stopped
	logger.D("checking abbot status")
	status, plainErr := ctr.State()
	if plainErr != nil {
		return fmt.Errorf("failed to get abbot contaienr status: %v", plainErr)
	}

	if status != define.ContainerStateRunning {
		ctx, cancel := r.ActionContext()
		defer cancel()

		logger.D("abbot not running, trying to start", log.String("status", status.String()))
		err := ctr.Start(ctx, true)
		if err != nil {
			return fmt.Errorf("failed to start abbot container: %v", err)
		}

		time.Sleep(5 * time.Second)
		logger.D("check abbot status again")
		status, plainErr = ctr.State()
		if plainErr != nil {
			return fmt.Errorf("failed to get abbot container status: %v", plainErr)
		}
		_ = status
	}

	cmd := append(ctr.Command(), subCmd...)
	logger.D("executing in abbot container", log.Strings("cmd", cmd))
	msgErr := r.execInContainer(ctr, nil, output, output, nil, cmd, false)
	if msgErr != nil {
		return fmt.Errorf("unable to execute network command: %s", msgErr.Description)
	}

	return nil
}

// InitRuntime will start all existing pods after runtime has been created
// if abbot container exists, start it first
// only fatal error will be returned
func (r *libpodRuntime) InitRuntime() error {
	logger := r.Log().WithFields(log.String("action", "init"))
	ctx, cancelInit := r.ActionContext()
	defer cancelInit()

	logger.D("looking up abbot container")
	abbotCtr, err2 := r.findAbbotContainer()
	if err2 == nil {
		logger.D("starting abbot container")
		_ = abbotCtr.Sync()
		status, err := abbotCtr.State()
		if err != nil {
			return fmt.Errorf("failed to inspect abbot container state: %v", err)
		}

		switch status {
		case define.ContainerStatePaused:
			err = abbotCtr.Unpause()
		case define.ContainerStateRunning:
			// do nothing
			err = nil
		default:
			err = abbotCtr.Start(ctx, true)
		}

		if err != nil {
			logger.I("failed to start abbot container", log.Error(err))
			// abbot found but failed to start
			return fmt.Errorf("failed to start abbot container: %v", err)
		}
	}

	logger.D("looking up all pods")
	allPods, err := r.runtimeClient.Pods()
	if err != nil {
		logger.I("failed to find all pods", log.Error(err))
		return err
	}
	for _, pod := range allPods {
		// only select valid pods
		if _, ok := pod.Labels()[constant.ContainerLabelPodUID]; !ok {
			logger.D("deleting invalid pod", log.String("name", pod.Name()))
			if err = r.deletePod(pod); err != nil {
				logger.I("failed to delete invalid pod", log.Error(err))
			}
		} else {
			logger.D("starting valid pod", log.String("name", pod.Name()))
			_, err = pod.Start(ctx)

			if err != nil {
				logger.I("failed to start pod", log.Error(err))
				continue
			}

			if runtimeutil.IsHostNetwork(pod.Labels()) {
				continue
			}

			containers, err := pod.AllContainers()
			if err != nil {
				logger.I("failed to list containers in pod", log.Error(err))
				continue
			}

			for _, ctr := range containers {
				if ctr.Labels()[constant.ContainerLabelPodContainerRole] != constant.ContainerRoleInfra {
					continue
				}

				_ = ctr.Sync()
				pid, _ := ctr.PID()
				err := r.networkClient.RestoreLink(ctr.ID(), uint32(pid))
				if err != nil {
					logger.I("failed to restore container network", log.Error(err))
				}
				break
			}
		}
	}

	return nil
}

// EnsureImages ensure container images
func (r *libpodRuntime) EnsureImages(options *aranyagopb.ImageEnsureCmd) ([]*aranyagopb.ImageStatusMsg, error) {
	logger := r.Log().WithFields(log.String("action", "ensureImages"), log.Any("options", options))
	logger.D("ensuring pod container image(s)")

	allImages := map[string]*aranyagopb.ImagePullConfig{
		r.PauseImage: {PullPolicy: aranyagopb.IMAGE_PULL_IF_NOT_PRESENT},
	}

	for imageName, opt := range options.Images {
		allImages[imageName] = opt
	}
	pulledImages, err := r.ensureImages(allImages)
	if err != nil {
		logger.I("failed to ensure container images", log.Error(err))
		return nil, err
	}

	var images []*aranyagopb.ImageStatusMsg
	for _, img := range pulledImages {
		var sha256Hash string
		digests, err := img.RepoDigests()
		if err != nil {
			return nil, err
		}
		for _, digest := range digests {
			idx := strings.LastIndex(digest, "sha256:")
			if idx > -1 {
				sha256Hash = digest[idx+7:]
			}
		}

		if sha256Hash == "" {
			continue
		}

		images = append(images, &aranyagopb.ImageStatusMsg{
			Sha256: sha256Hash,
			Names:  []string{img.Tag},
		})
	}

	return images, nil
}

// CreateContainers creates containers
// nolint:gocyclo
func (r *libpodRuntime) CreateContainers(options *aranyagopb.PodEnsureCmd) (_ *aranyagopb.PodStatusMsg, err error) {
	logger := r.Log().WithFields(log.String("action", "createContainers"), log.String("uid", options.PodUid))
	ctx, cancelCreate := r.RuntimeActionContext()
	defer func() {
		cancelCreate()

		if err != nil {
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
		pod      *libpod.Pod
		pauseCtr *libpod.Container
		podIP    string
	)

	logger.V("looking up previously created pod")
	pod, err = r.findPod(options.PodUid)
	if err != nil {
		if !errors.Is(err, wellknownerrors.ErrNotFound) {
			return nil, fmt.Errorf("failed to look up pod: %w", err)
		}

		err = nil
		// need to create pause container
		logger.V("creating pod and pause container")
		pod, pauseCtr, podIP, err = r.createPauseContainer(ctx, options)
		if err != nil {
			return nil, fmt.Errorf("failed to create pod for containers: %w", err)
		}
	} else {
		// pod is there, pause container should present
		logger.V("found pod, check pause container")
		pauseCtr, err = r.findContainer(options.PodUid, constant.ContainerNamePause)
		if err != nil {
			logger.I("deleting invalid pod without pause container")
			if err2 := r.deletePod(pod); err2 != nil {
				logger.I("failed to delete invalid pod", log.Error(err2))
			}

			return nil, err
		}
	}

	defer func() {
		if err != nil {
			logger.D("deleting pause container due to error", log.Error(err))
			err2 := r.deletePod(pod)
			if err2 != nil {
				logger.I("failed to delete pause container", log.Error(err2))
			}
		}
	}()

	logger.V("starting pause container")
	_ = pauseCtr.Sync()
	state, err := pauseCtr.State()
	if err != nil {
		return nil, fmt.Errorf("failed to get pause container state: %w", err)
	}

	switch state {
	case define.ContainerStateConfigured,
		define.ContainerStateCreated,
		define.ContainerStateStopped,
		define.ContainerStateExited:
		// can be started
		err = pauseCtr.Start(ctx, false)
		if err != nil {
			return nil, fmt.Errorf("failed to start pause container: %w", err)
		}
	case define.ContainerStatePaused:
		err = pauseCtr.Unpause()
		if err != nil {
			return nil, fmt.Errorf("failed to unpause pause container: %w", err)
		}
	case define.ContainerStateUnknown:
		return nil, fmt.Errorf("pause container state unknow")
	case define.ContainerStateRemoving:
		return nil, fmt.Errorf("pause container is to be removed")
	case define.ContainerStateRunning:
	// already running (do nothing)
	default:
		return nil, fmt.Errorf("invalid container state %v", state)
	}

	var (
		ctrList           []*libpod.Container
		ctrPostStartHooks []*aranyagopb.ActionMethod
	)

	for _, spec := range options.Containers {
		logger.V("creating container", log.String("name", spec.Name))

		var ctr *libpod.Container
		ctr, err = r.createContainer(ctx, options, pod, spec, runtimeutil.SharedNamespaces(pauseCtr.ID(), options))
		if err != nil {
			return nil, fmt.Errorf("failed to create container: %w", err)
		}

		ctrList = append(ctrList, ctr)
		ctrPostStartHooks = append(ctrPostStartHooks, spec.HookPostStart)
	}

	// start containers one by one
	for i, ctr := range ctrList {
		err = ctr.Start(ctx, false)
		if err != nil {
			return nil, fmt.Errorf("failed to start container: %w", err)
		}

		//noinspection GoNilness
		if h := ctrPostStartHooks[i]; h != nil {
			err = r.doHookAction(logger, ctr, h)
			if err != nil {
				return nil, fmt.Errorf("failed to execute post start hooks: %w", err)
			}
		}

		if !options.Wait {
			continue
		}

		var exitCode int32
		exitCode, err = ctr.WaitWithInterval(500 * time.Millisecond)
		if err != nil {
			return nil, fmt.Errorf("container exited with error code %d: %w", exitCode, err)
		}
	}

	time.Sleep(5 * time.Second)
	for _, ctr := range ctrList {
		if options.Wait {
			// container MUST have exited, no more check
			continue
		}

		logger.V("checking whether container exited")
		code, exited, err := ctr.ExitCode()
		if err != nil {
			return nil, fmt.Errorf("failed to check whether container %q exited: %w", ctr.Name(), err)
		}

		if exited {
			return nil, fmt.Errorf("container %q exited with code %d: %w", ctr.Name(), code, err)
		}
	}

	return r.translatePodStatus(podIP, pauseCtr, ctrList)
}

func (r *libpodRuntime) DeleteContainers(podUID string, containers []string) (*aranyagopb.PodStatusMsg, error) {
	logger := r.Log().WithFields(log.String("action", "deleteContainer"))
	pod, err := r.findPod(podUID)
	if err != nil {
		return nil, fmt.Errorf("failed to find pod: %w", err)
	}

	for _, ctrName := range containers {
		logger.V("looking up container", log.String("name", ctrName))
		ctr, err := r.findContainer(podUID, ctrName)
		if err != nil && !errors.Is(err, wellknownerrors.ErrNotFound) {
			return nil, fmt.Errorf("failed to do container find: %w", err)
		}

		logger.V("deleting container", log.String("name", ctrName))
		err = r.deleteContainer(ctr)
		if err != nil {
			return nil, fmt.Errorf("failed to delete container %q: %w", ctrName, err)
		}
	}

	var pauseCtr *libpod.Container
	ctrs, _ := pod.AllContainers()
	for _, c := range ctrs {
		if runtimeutil.IsPauseContainer(c.Labels()) {
			pauseCtr = c
			break
		}
	}

	if pauseCtr == nil {
		return nil, fmt.Errorf("no pause container in pod")
	}

	return r.translatePodStatus("", pauseCtr, ctrs)
}

func (r *libpodRuntime) DeletePod(options *aranyagopb.PodDeleteCmd) (*aranyagopb.PodStatusMsg, error) {
	logger := r.Log().WithFields(log.String("action", "deletePod"), log.Any("options", options))

	defer func() {
		logger.D("cleaning up pod data")
		err := runtimeutil.CleanupPodData(
			r.PodDir(options.PodUid),
			r.PodRemoteVolumeDir(options.PodUid, ""),
			r.PodTmpfsVolumeDir(options.PodUid, ""),
			r.storage,
		)
		if err != nil {
			logger.E("failed to cleanup pod data", log.Error(err))
		}
	}()

	logger.D("looking up pod to be deleted")
	pods, err := r.runtimeClient.Pods(podLabelFilterFunc(map[string]string{
		constant.ContainerLabelPodUID: options.PodUid,
	}))
	if err != nil {
		logger.I("failed to find the pod to be deleted", log.Error(err))
		return nil, err
	}

	for _, pod := range pods {
		logger.D("deleting target pod")
		if err := r.deletePod(pod); err != nil {
			logger.I("failed to delete target pod", log.Error(err))
			return nil, err
		}
	}

	return aranyagopb.NewPodStatusMsg(options.PodUid, "", nil), nil
}

func (r *libpodRuntime) ListPods(options *aranyagopb.PodListCmd) ([]*aranyagopb.PodStatusMsg, error) {
	logger := r.Log().WithFields(log.String("action", "list"), log.Any("options", options))
	filters := make(map[string]string)
	if !options.All {
		if options.Namespace != "" {
			filters[constant.ContainerLabelPodNamespace] = options.Namespace
		}

		if options.Name != "" {
			filters[constant.ContainerLabelPodName] = options.Name
		}
	}

	logger.D("listing pods")
	pods, err := r.runtimeClient.Pods(podLabelFilterFunc(filters))
	if err != nil {
		logger.I("failed to filter out desired pods")
		return nil, err
	}

	actionCtx, cancel := r.RuntimeActionContext()
	defer cancel()

	var results []*aranyagopb.PodStatusMsg
	for _, pod := range pods {
		podUID, ok := pod.Labels()[constant.ContainerLabelPodUID]
		if !ok {
			logger.D("deleting invalid pod", log.String("name", pod.Name()))
			if err := r.deletePod(pod); err != nil {
				logger.I("failed to delete invalid pod", log.Error(err))
			}
			continue
		}

		logger.D("looking up pause container", log.String("podUID", podUID))
		pauseCtr, err := r.findContainer(podUID, constant.ContainerNamePause)
		if err != nil {
			// in libpod, pod is separated from the containers, it's highly possible we didn't
			// create any container in this pod, so just delete it silently
			err2 := r.runtimeClient.RemovePod(actionCtx, pod, true, true)
			if err2 != nil {
				logger.I("failed to remove invalid pod", log.Error(err2))
				return nil, err2
			}
			continue
		}

		var podIP string
		// force sync
		_ = pauseCtr.Sync()
		if runtimeutil.IsHostNetwork(pauseCtr.Labels()) {
			logger.D("looking up pod ip for non-host network pod")
			pid, _ := pauseCtr.PID()
			podIP, err = r.networkClient.GetAddress(uint32(pid))
			if err != nil {
				logger.I("failed to get pod ip address", log.Error(err))
			}
		}

		logger.D("looking up all pod containers", log.String("podUID", podUID))
		ctrList, err := pod.AllContainers()
		if err != nil {
			return nil, fmt.Errorf("failed to find all containers in pod: %w", err)
		}

		logger.D("translating pod status", log.String("podUID", podUID))
		status, err := r.translatePodStatus(podIP, pauseCtr, ctrList)
		if err != nil {
			return nil, fmt.Errorf("failed to translate pod status: %w", err)
		}

		results = append(results, status)
	}

	return results, nil
}

func (r *libpodRuntime) ExecInContainer(
	podUID, container string,
	stdin io.Reader,
	stdout, stderr io.Writer,
	resizeCh <-chan *aranyagopb.ContainerTerminalResizeCmd,
	command []string,
	tty bool,
) *aranyagopb.ErrorMsg {
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

	return r.execInContainer(ctr, stdin, stdout, stderr, resizeCh, command, tty)
}

func (r *libpodRuntime) AttachContainer(
	podUID, container string,
	stdin io.Reader,
	stdout, stderr io.Writer,
	resizeCh <-chan *aranyagopb.ContainerTerminalResizeCmd,
) error {
	logger := r.Log().WithFields(
		log.String("action", "attach"),
		log.String("uid", podUID),
		log.String("container", container),
	)
	logger.D("attach to pod container")

	ctr, err := r.findContainer(podUID, container)
	if err != nil {
		logger.I("failed to find container", log.Error(err))
		return fmt.Errorf("failed to find target container: %w", wellknownerrors.ErrNotFound)
	}

	attachCtx, cancelAttach := r.ActionContext()
	defer cancelAttach()

	var ch chan remotecommand.TerminalSize
	if resizeCh != nil {
		ch = make(chan remotecommand.TerminalSize, 1)
		go func() {
			for {
				select {
				case size, more := <-resizeCh:
					if !more {
						return
					}
					select {
					case <-attachCtx.Done():
					case ch <- remotecommand.TerminalSize{Width: uint16(size.Cols), Height: uint16(size.Rows)}:
					}
				case <-attachCtx.Done():
				}
			}

		}()
	}

	streams := r.translateStreams(stdin, stdout, stderr)

	err = ctr.Attach(streams, "", ch)
	if err != nil {
		return err
	}

	return nil
}

func (r *libpodRuntime) GetContainerLogs(
	podUID string,
	options *aranyagopb.ContainerLogsCmd,
	stdout, stderr io.WriteCloser,
	logCtx context.Context,
) error {
	defer func() { _, _ = stdout.Close(), stderr.Close() }()

	ctr, err := r.findContainer(podUID, options.Container)
	if err != nil {
		return fmt.Errorf("failed to find target container")
	}

	err = runtimeutil.ReadLogs(logCtx, ctr.LogPath(), options, stdout, stderr)
	if err != nil {
		return err
	}

	return nil
}

func (r *libpodRuntime) PortForward(
	podUID, protocol string,
	port int32,
	downstream io.ReadWriter,
) error {
	logger := r.Log().WithFields(
		log.String("action", "portforward"),
		log.String("proto", protocol),
		log.Int32("port", port),
		log.String("uid", podUID),
	)
	logger.D("port-forwarding to pod container")

	pauseCtr, err := r.findContainer(podUID, constant.ContainerNamePause)
	if err != nil {
		logger.I("failed to find pause container", log.Error(err))
		return wellknownerrors.ErrNotFound
	}

	addresses, plainErr := pauseCtr.IPs()
	if plainErr != nil {
		logger.I("failed to get container ips", log.Error(err))
	}

	var address string
	for _, addr := range addresses {
		if !addr.IP.IsLoopback() {
			address = addr.String()
		}
	}

	ctx, cancel := r.ActionContext()
	defer cancel()

	return runtimeutil.PortForward(ctx, address, protocol, port, downstream)
}

func (r *libpodRuntime) UpdateContainerNetwork(options *aranyagopb.NetworkUpdatePodNetworkCmd) ([]*aranyagopb.PodStatusMsg, error) {
	logger := r.Log().WithFields(log.String("action", "updateContainerNetwork"))

	logger.D("looking up abbot container")
	_, err := r.findAbbotContainer()
	if err != nil {
		return nil, fmt.Errorf("failed to find abbot container: %w", err)
	}

	logger.D("inspecting all pods for container network update")
	pods, err := r.runtimeClient.GetAllPods()
	if err != nil {
		return nil, fmt.Errorf("failed to get all pods: %w", err)
	}

	var result []*aranyagopb.PodStatusMsg
	for _, pod := range pods {
		// only select valid pods
		if podUID, ok := pod.Labels()[constant.ContainerLabelPodUID]; ok {
			logger.I("looking up pause container in pod", log.String("podUID", podUID))
			pauseCtr, err := r.findContainer(podUID, constant.ContainerNamePause)
			if err != nil {
				return nil, fmt.Errorf("failed to find pause container: %w", err)
			}

			_ = pauseCtr.Sync()
			status, _ := pauseCtr.State()
			if status == define.ContainerStateRunning && !runtimeutil.IsHostNetwork(pauseCtr.Labels()) {
				pid, _ := pauseCtr.PID()
				logger.D("ensuring pod address", log.String("name", pauseCtr.Name()))
				podIP, err := r.networkClient.EnsureAddress(pauseCtr.ID(), uint32(pid), options)
				if err != nil {
					return nil, fmt.Errorf("failed to ensure pod address: %w", err)
				}
				result = append(result, aranyagopb.NewPodStatusMsg(podUID, podIP, nil))
			}
		}
	}

	return result, nil
}
