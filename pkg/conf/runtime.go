package conf

import (
	"path/filepath"
	"time"

	"github.com/spf13/pflag"

	"arhat.dev/arhat/pkg/constant"
)

type RuntimeEndpoint struct {
	Endpoint      string        `json:"endpoint" yaml:"endpoint"`
	DialTimeout   time.Duration `json:"dialTimeout" yaml:"dialTimeout"`
	ActionTimeout time.Duration `json:"actionTimeout" yaml:"actionTimeout"`
}

func flagsForArhatRuntimeEndpointConfig(prefix string, config *RuntimeEndpoint) *pflag.FlagSet {
	fs := pflag.NewFlagSet("arhat.runtime.endpoint", pflag.ExitOnError)

	fs.StringVar(&config.Endpoint, prefix+"endpoint", "", "set endpoint address")
	fs.DurationVar(&config.DialTimeout, prefix+"dialTimeout",
		constant.DefaultRuntimeDialTimeout, "set endpoint dial timeout")
	fs.DurationVar(&config.ActionTimeout, prefix+"actionTimeout",
		constant.DefaultRuntimeActionTimeout, "set endpoint maximum action timeout")

	return fs
}

type RuntimeConfig struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	DataDir string `json:"dataDir" yaml:"dataDir"`

	// pause image and command
	PauseImage   string `json:"pauseImage" yaml:"pauseImage"`
	PauseCommand string `json:"pauseCommand" yaml:"pauseCommand"`

	// ManagementNamespace the name used to separate user view scope
	ManagementNamespace string `json:"managementNamespace" yaml:"managementNamespace"`

	// Optional
	EndPoints struct {
		// image endpoint
		Image RuntimeEndpoint `json:"image" yaml:"image"`
		// runtime endpoint
		Container RuntimeEndpoint `json:"container" yaml:"container"`
	} `json:"endpoints" yaml:"endpoints"`
}

func FlagsForArhatRuntimeConfig(prefix string, config *RuntimeConfig) *pflag.FlagSet {
	fs := pflag.NewFlagSet("arhat.runtime.endpoint", pflag.ExitOnError)

	fs.BoolVar(&config.Enabled, prefix+"enabled", true, "enable runtime or use none runtime")
	fs.StringVar(&config.DataDir, prefix+"dataDir", constant.DefaultArhatDataDir, "set runtime data root dir")
	fs.StringVar(&config.PauseImage, prefix+"pauseImage", constant.DefaultPauseImage, "set pause image")
	fs.StringVar(&config.PauseCommand, prefix+"pauseCommand",
		constant.DefaultPauseCommand, "set pause image command")
	fs.StringVar(&config.ManagementNamespace, prefix+"managementNamespace",
		constant.DefaultManagementNamespace, "set container management namespace (for libpod)")

	fs.AddFlagSet(flagsForArhatRuntimeEndpointConfig(prefix+"endpoints.image.", &config.EndPoints.Image))
	fs.AddFlagSet(flagsForArhatRuntimeEndpointConfig(prefix+"endpoints.container.", &config.EndPoints.Container))

	return fs
}

func (c *RuntimeConfig) PodDir(podUID string) string {
	return filepath.Join(c.DataDir, "pods", podUID)
}

func (c *RuntimeConfig) podVolumeDir(podUID, typ, volumeName string) string {
	return filepath.Join(c.PodDir(podUID), "volumes", typ, volumeName)
}

func (c *RuntimeConfig) PodRemoteVolumeDir(podUID, volumeName string) string {
	return c.podVolumeDir(podUID, "remote", volumeName)
}

func (c *RuntimeConfig) PodBindVolumeDir(podUID, volumeName string) string {
	return c.podVolumeDir(podUID, "bind", volumeName)
}

func (c *RuntimeConfig) PodTmpfsVolumeDir(podUID, volumeName string) string {
	return c.podVolumeDir(podUID, "tmpfs", volumeName)
}

func (c *RuntimeConfig) PodResolvConfFile(podUID string) string {
	return filepath.Join(c.PodDir(podUID), "volumes", "bind", "_net", "resolv.conf")
}
