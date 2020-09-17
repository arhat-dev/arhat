// +build !rt_docker,!rt_none

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

package main

import (
	"arhat.dev/arhat/pkg/constant"
	"github.com/spf13/pflag"
)

func setPlatformSpecificRuntimeFlags(flags *pflag.FlagSet) {
	flags.VisitAll(func(f *pflag.Flag) {
		switch f.Name {
		case "runtime.endpoints.container.endpoint":
			f.DefValue = constant.DefaultDockerWindowsEndpoint
		case "runtime.endpoints.image.endpoint":
			f.DefValue = constant.DefaultDockerWindowsEndpoint
		}
	})
}
