// +build !noclient_coap

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

package coap

import (
	"arhat.dev/arhat/pkg/client/clientutil"
	"arhat.dev/arhat/pkg/conf"
)

type Config struct {
	clientutil.CommonConfig `json:",inline" yaml:",inline"`

	PathNamespaceFrom conf.ValueFromSpec `json:"pathNamespaceFrom" yaml:"pathNamespaceFrom"`
	Transport         string             `json:"transport" yaml:"transport"`
	URIQueries        map[string]string  `json:"uriQueries" yaml:"uriQueries"`
	Keepalive         int32              `json:"keepalive" yaml:"keepalive"`
}
