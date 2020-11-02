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

package constant

import "time"

const (
	// arhat defaults
	DefaultArhatConfigFile     = "/etc/arhat/config.yaml"
	DefaultArhatDataDir        = "/var/lib/arhat/data"
	DefaultPauseImage          = "k8s.gcr.io/pause:3.1"
	DefaultPauseCommand        = "/pause"
	DefaultManagementNamespace = "container.arhat.dev"
	DefaultProcfsPath          = "/proc"
	DefaultSysfsPath           = "/sys"

	DefaultDockerUnixEndpoint    = "unix:///var/run/docker.sock"
	DefaultDockerWindowsEndpoint = "tcp://docker.for.win.localhost:2375"
	DefaultRuntimeDialTimeout    = time.Minute
	DefaultRuntimeActionTimeout  = 10 * time.Minute
)

// Extension defaults
const (
	// peripheral
	DefaultPeripheralMetricsMaxCacheTime = 30 * time.Minute
)

const (
	DefaultInteractiveStreamReadTimeout    = 20 * time.Millisecond
	DefaultNonInteractiveStreamReadTimeout = 200 * time.Millisecond
	DefaultPortForwardStreamReadTimeout    = 50 * time.Millisecond
)

const (
	DefaultExitCodeOnError = 128
)
