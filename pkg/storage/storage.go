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

package storage

import (
	"context"
	"fmt"

	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/arhat/pkg/constant"
	"arhat.dev/arhat/pkg/types"
)

func NewStorage(appCtx context.Context, config *conf.StorageConfig) (types.Storage, error) {
	switch config.Backend {
	case constant.StorageBackendSSHFS:
		return NewSSHFSStorage(appCtx, config)
	case "":
		return NewNoneStorage()
	}

	return nil, fmt.Errorf("unknown storage backend %s", config.Backend)
}
