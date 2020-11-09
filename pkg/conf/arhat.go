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

package conf

import (
	"context"
	"io/ioutil"

	"arhat.dev/pkg/exechelper"
	"arhat.dev/pkg/log"
	"ext.arhat.dev/runtimeutil/storage"
	"github.com/spf13/pflag"
)

// Config
type Config struct {
	Arhat        AppConfig            `json:"arhat" yaml:"arhat"`
	Connectivity ConnectivityConfig   `json:"connectivity" yaml:"connectivity"`
	Storage      storage.ClientConfig `json:"storage" yaml:"storage"`
	Extension    ExtensionConfig      `json:"extension" yaml:"extension"`
}

// AppConfig configuration for arhat application behavior
type AppConfig struct {
	Log log.ConfigSet `json:"log" yaml:"log"`

	Chroot string `json:"chroot" yaml:"chroot"`

	Host HostConfig `json:"host" yaml:"host"`
	Node NodeConfig `json:"node" yaml:"node"`

	Optimization struct {
		MaxProcessors int `json:"maxProcessors" yaml:"maxProcessors"`
	} `json:"optimization" yaml:"optimization"`
}

type HostConfig struct {
	AllowAttach      bool `json:"allowAttach" yaml:"allowAttach"`
	AllowExec        bool `json:"allowExec" yaml:"allowExec"`
	AllowLog         bool `json:"allowLog" yaml:"allowLog"`
	AllowPortForward bool `json:"allowPortForward" yaml:"allowPortForward"`
}

func FlagsForArhatHostConfig(prefix string, config *HostConfig) *pflag.FlagSet {
	fs := pflag.NewFlagSet("arhat.host", pflag.ExitOnError)

	fs.BoolVar(&config.AllowAttach, prefix+"allowAttach", false, "allow kubectl attach")
	fs.BoolVar(&config.AllowExec, prefix+"allowExec", false, "allow kubectl exec")
	fs.BoolVar(&config.AllowLog, prefix+"allowLog", false, "allow kubectl logs")
	fs.BoolVar(&config.AllowPortForward, prefix+"allowPortForward", false, "allow kubectl port-forward")

	return fs
}

type NodeConfig struct {
	MachineIDFrom ValueFromSpec `json:"machineIDFrom" yaml:"machineIDFrom"`
	ExtInfo       []NodeExtInfo `json:"extInfo" yaml:"extInfo"`
}

type ValueFromSpec struct {
	Exec []string `json:"exec" yaml:"exec"`
	File string   `json:"file" yaml:"file"`
	Text string   `json:"text" yaml:"text"`
}

func (vf *ValueFromSpec) Get() (string, error) {
	if vf == nil {
		return "", nil
	}

	if vf.Text != "" {
		return vf.Text, nil
	}

	if vf.File != "" {
		data, err := ioutil.ReadFile(vf.File)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	if len(vf.Exec) > 0 {
		cmd, err := exechelper.Prepare(context.TODO(), vf.Exec, nil, false, nil)
		if err != nil {
			return "", err
		}

		data, err := cmd.CombinedOutput()
		if err != nil {
			return "", err
		}

		return string(data), nil
	}

	return "", nil
}

type NodeExtInfo struct {
	ValueFrom ValueFromSpec `json:"valueFrom" yaml:"valueFrom"`

	ValueType string `json:"valueType" yaml:"valueType"`
	Operator  string `json:"operator" yaml:"operator"`
	ApplyTo   string `json:"applyTo" yaml:"applyTo"`
}
