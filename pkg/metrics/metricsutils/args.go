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

package metricsutils

import (
	"fmt"
	"strings"
)

func GetEnabledCollectors(list []string) map[string]struct{} {
	enabledCollectors := make(map[string]struct{})
	for _, c := range list {
		for _, coll := range strings.Split(c, ",") {
			enabledCollectors[coll] = struct{}{}
		}
	}

	return enabledCollectors
}

func GetExtraArgs(enabledCollectors map[string]struct{}, args []string) ([]string, error) {
	var result []string
	for _, a := range args {
		parts := strings.SplitN(strings.SplitN(a, "=", 2)[0], ".", 3)
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid arg %q", a)
		}

		if _, ok := enabledCollectors[parts[1]]; !ok {
			return nil, fmt.Errorf("collector %q not enabled", parts[1])
		}

		result = append(result, a)
	}

	return result, nil
}
