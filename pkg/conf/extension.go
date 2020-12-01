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
	"time"

	"arhat.dev/pkg/tlshelper"
	"github.com/spf13/pflag"

	"arhat.dev/arhat/pkg/constant"
)

// nolint:maligned
type ExtensionConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`

	Endpoints  []ExtensionEndpoint       `json:"endpoints" yaml:"endpoints"`
	Peripheral PeripheralExtensionConfig `json:"peripheral" yaml:"peripheral"`
	Runtime    RuntimeExtensionConfig    `json:"runtime" yaml:"runtime"`
}

type ExtensionEndpoint struct {
	Listen string `json:"listen" yaml:"listen"`
	TLS    struct {
		tlshelper.TLSConfig `json:",inline" yaml:",inline"`
		VerifyClientCert    bool `json:"verifyClientCert" yaml:"verifyClientCert"`
	} `json:"tls" yaml:"tls"`

	KeepaliveInterval time.Duration `json:"keepaliveInterval" yaml:"keepaliveInterval"`
	MessageTimeout    time.Duration `json:"messageTimeout" yaml:"messageTimeout"`
}

type PeripheralExtensionConfig struct {
	MetricsCacheTimeout time.Duration `json:"metricsCacheTimeout" yaml:"metricsCacheTimeout"`
}

type RuntimeExtensionConfig struct {
	Wait bool `json:"wait" yaml:"wait"`
}

func FlagsForExtensionConfig(prefix string, config *ExtensionConfig) *pflag.FlagSet {
	fs := pflag.NewFlagSet("extension", pflag.ExitOnError)

	fs.BoolVar(&config.Enabled, prefix+"enable", false, "enable extension server")

	fs.DurationVar(&config.Peripheral.MetricsCacheTimeout, prefix+"metricsCacheTimeout",
		constant.DefaultPeripheralMetricsCacheTimeout, "peripheral metrics cache timeout")

	return fs
}
