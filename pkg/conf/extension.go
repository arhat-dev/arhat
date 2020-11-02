package conf

import (
	"time"

	"arhat.dev/pkg/confhelper"
	"github.com/spf13/pflag"

	"arhat.dev/arhat/pkg/constant"
)

type ExtensionConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`

	Endpoints   []ExtensionEndpoint       `json:"endpoints" yaml:"endpoints"`
	Peripherals PeripheralExtensionConfig `json:"peripherals" yaml:"peripherals"`
}

type ExtensionEndpoint struct {
	Listen string `json:"listen" yaml:"listen"`
	TLS    struct {
		confhelper.TLSConfig `json:",inline" yaml:",inline"`
		VerifyClientCert     bool `json:"verifyClientCert" yaml:"verifyClientCert"`
	} `json:"tls" yaml:"tls"`

	KeepaliveInterval time.Duration `json:"keepaliveInterval" yaml:"keepaliveInterval"`
	MessageTimeout    time.Duration `json:"messageTimeout" yaml:"messageTimeout"`
}

type PeripheralExtensionConfig struct {
	MetricsCacheTimeout time.Duration `json:"metricsCacheTimeout" yaml:"metricsCacheTimeout"`
}

func FlagsForExtensionConfig(prefix string, config *ExtensionConfig) *pflag.FlagSet {
	fs := pflag.NewFlagSet("extension", pflag.ExitOnError)

	fs.BoolVar(&config.Enabled, prefix+"enable", false, "enable extension server")

	fs.DurationVar(&config.Peripherals.MetricsCacheTimeout, prefix+"metricsCacheTimeout",
		constant.DefaultPeripheralMetricsCacheTimeout, "peripheral metrics cache timeout")

	return fs
}
