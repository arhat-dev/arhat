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

package util

import (
	"arhat.dev/arhat-proto/arhatgopb"

	"arhat.dev/libext/types"
)

func NewMsg(
	marshal types.MarshalFunc,
	kind arhatgopb.MsgType,
	id, ack uint64, body interface{},
) (*arhatgopb.Msg, error) {
	payload, err := marshal(body)
	if err != nil {
		return nil, err
	}

	return &arhatgopb.Msg{
		Kind:    kind,
		Id:      id,
		Ack:     ack,
		Payload: payload,
	}, nil
}

func GetMsgType(m interface{}) arhatgopb.MsgType {
	switch m.(type) {
	case *arhatgopb.RegisterMsg:
		return arhatgopb.MSG_REGISTER
	case *arhatgopb.PeripheralOperationResultMsg:
		return arhatgopb.MSG_PERIPHERAL_OPERATION_RESULT
	case *arhatgopb.PeripheralMetricsMsg:
		return arhatgopb.MSG_PERIPHERAL_METRICS
	case *arhatgopb.DoneMsg:
		return arhatgopb.MSG_DONE
	case *arhatgopb.ErrorMsg:
		return arhatgopb.MSG_ERROR
	case *arhatgopb.PeripheralEventMsg:
		return arhatgopb.MSG_PERIPHERAL_EVENTS
	default:
		return 0
	}
}
