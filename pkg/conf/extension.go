package conf

import (
	"arhat.dev/pkg/confhelper"
	"time"
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
