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

package protoutil

import (
	"arhat.dev/arhat-proto/arhatgopb"

	"arhat.dev/libext/codec"
)

func NewCmd(
	marshal codec.MarshalFunc,
	kind arhatgopb.CmdType,
	id, seq uint64, body interface{},
) (*arhatgopb.Cmd, error) {
	payload, err := marshal(body)
	if err != nil {
		return nil, err
	}

	return &arhatgopb.Cmd{
		Kind:    kind,
		Id:      id,
		Seq:     seq,
		Payload: payload,
	}, nil
}
