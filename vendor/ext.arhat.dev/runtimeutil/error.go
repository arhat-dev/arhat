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

package runtimeutil

import "fmt"

func CollectErrors(errCh <-chan error) error {
	var errAll error
	for err := range errCh {
		if err != nil {
			if errAll != nil {
				errAll = fmt.Errorf("%s; %w", errAll.Error(), err)
			} else {
				errAll = err
			}
		}
	}

	return errAll
}
