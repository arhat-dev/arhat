package conf

import (
	"time"

	"arhat.dev/pkg/confhelper"
	"github.com/spf13/pflag"

	"arhat.dev/arhat/pkg/constant"
)

type ExtensionConfig struct {
	Enabled bool                 `json:"enabled" yaml:"enabled"`
	Listen  string               `json:"listen" yaml:"listen"`
	TLS     confhelper.TLSConfig `json:"tls" yaml:"tls"`

	Devices DeviceExtensionConfig `json:"devices" yaml:"devices"`
}

type DeviceExtensionConfig struct {
	MaxMetricsCacheTime time.Duration `json:"maxMetricsCacheTime" yaml:"maxMetricsCacheTime"`
}

func FlagsForExtensionConfig(prefix string, config *ExtensionConfig) *pflag.FlagSet {
	fs := pflag.NewFlagSet("extension", pflag.ExitOnError)

	fs.BoolVar(&config.Enabled, prefix+"enable",
		false, "enable extension server")
	fs.StringVar(&config.Listen, prefix+"listen",
		constant.DefaultArhatExtensionListen, "extension server listen address")
	fs.AddFlagSet(confhelper.FlagsForTLSConfig(prefix+"tls.", &config.TLS))

	fs.DurationVar(&config.Devices.MaxMetricsCacheTime, prefix+"maxMetricsCacheTime",
		constant.DefaultDeviceMetricsMaxCacheTime, "device metrics cache timeout")

	return fs
}
