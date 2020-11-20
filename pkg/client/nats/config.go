// +build !noclient_nats

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

package nats

import (
	"time"

	"arhat.dev/arhat/pkg/client/clientutil"
)

type Config struct {
	clientutil.CommonConfig `json:",inline" yaml:",inline"`

	PingInterval time.Duration `json:"pingInterval" yaml:"pingInterval"`

	SubjectNamespace string `json:"subjectNamespace" yaml:"subjectNamespace"`

	// ClusterID of the nats streaming server
	ClusterID string `json:"clusterID" yaml:"clusterID"`

	// ClientID of the nats streaming client, will also be used as nats
	// client Name (which is an optional name label sent to the server
	// on CONNECT to identify the client)
	ClientID string `json:"clientID" yaml:"clientID"`

	MaxPendingPubAck int `json:"maxPendingPubAck" yaml:"maxPendingPubAck"`

	// AckWait time to wait before resend message (both client and server)
	AckWait time.Duration `json:"ackWait" yaml:"ackWait"`

	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	Token    string `json:"token" yaml:"token"`

	NKey string `json:"nkey" yaml:"nkey"`
	JWT  string `json:"jwt" yaml:"jwt"`
}
