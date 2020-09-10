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
	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/wellknownerrors"

	"arhat.dev/arhat/pkg/types"
)

func NewNoneStorage() (types.Storage, error) {
	return new(noneStorage), nil
}

type noneStorage struct{}

func (s *noneStorage) Name() string {
	return "none"
}

func (s *noneStorage) Mount(remotePath, mountPoint string, onExited types.StorageFailureHandleFunc) error {
	return wellknownerrors.ErrNotSupported
}

func (s *noneStorage) Unmount(mountPoint string) error {
	return wellknownerrors.ErrNotSupported
}

func (s *noneStorage) SetCredentials(options *aranyagopb.StorageCredentialOptions) {
}
