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
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"arhat.dev/aranya-proto/aranyagopb/aranyagoconst"
	"arhat.dev/pkg/confhelper"
	"arhat.dev/pkg/exechelper"
	"arhat.dev/pkg/log"
	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/pflag"

	"arhat.dev/arhat/pkg/constant"
)

// ArhatConfig
type ArhatConfig struct {
	Arhat        ArhatAppConfig          `json:"arhat" yaml:"arhat"`
	Connectivity ArhatConnectivityConfig `json:"connectivity" yaml:"connectivity"`
	Runtime      ArhatRuntimeConfig      `json:"runtime" yaml:"runtime"`
	Storage      ArhatStorageConfig      `json:"storage" yaml:"storage"`
}

func (c *ArhatConfig) GetLogConfig() log.ConfigSet {
	return c.Arhat.Log
}

func (c *ArhatConfig) SetLogConfig(l log.ConfigSet) {
	c.Arhat.Log = l
}

// ArhatConnectivityConfig configuration for connectivity part in arhat
type ArhatConnectivityConfig struct {
	DialTimeout    time.Duration `json:"dialTimeout" yaml:"dialTimeout"`
	InitialBackoff time.Duration `json:"initialBackoff" yaml:"initialBackoff"`
	MaxBackoff     time.Duration `json:"maxBackoff" yaml:"maxBackoff"`
	BackoffFactor  float64       `json:"backoffFactor" yaml:"backoffFactor"`

	ArhatConnectivityMethods `json:",inline" yaml:",inline"`
}

type ArhatConnectivityMethods struct {
	Method     string          `json:"method" yaml:"method"`
	MQTTConfig ArhatMQTTConfig `json:"mqtt" yaml:"mqtt"`
	GRPCConfig ArhatGRPCConfig `json:"grpc" yaml:"grpc"`
	CoAPConfig ArhatCoAPConfig `json:"coap" yaml:"coap"`
}

func FlagsForArhatConnectivityConfig(prefix string, config *ArhatConnectivityConfig) *pflag.FlagSet {
	fs := pflag.NewFlagSet("arhat.conn", pflag.ExitOnError)

	fs.StringVar(&config.Method, prefix+"method", "", "set connectivity method")
	fs.DurationVar(&config.DialTimeout, prefix+"dialTimeout", 10*time.Second,
		"set dial timeout for connectivity")
	fs.DurationVar(&config.InitialBackoff, prefix+"initialBackoff", 1*time.Second,
		"set initial connect backoff timeout")
	fs.DurationVar(&config.MaxBackoff, prefix+"maxBackoff", 30*time.Second,
		"set max connect backoff timeout")
	fs.Float64Var(&config.BackoffFactor, prefix+"backoffFactor", 1.2, "set backoff factor")

	fs.AddFlagSet(flagsForArhatMQTTConnectivityConfig(prefix+"mqtt.", &config.MQTTConfig))
	fs.AddFlagSet(flagsForArhatGRPCConnectivityConfig(prefix+"grpc.", &config.GRPCConfig))
	fs.AddFlagSet(flagsForArhatCoAPConnectivityConfig(prefix+"coap.", &config.CoAPConfig))

	return fs
}

type arhatConnectivityCommonConfig struct {
	Endpoint string               `json:"endpoint" yaml:"endpoint"`
	Priority int                  `json:"priority" yaml:"priority"`
	TLS      confhelper.TLSConfig `json:"tls" yaml:"tls"`
}

func flagsForArhatConnectivityCommonConfig(prefix string, config *arhatConnectivityCommonConfig) *pflag.FlagSet {
	fs := pflag.NewFlagSet("arhat.conn._common", pflag.ExitOnError)

	fs.StringVar(&config.Endpoint, prefix+"endpoint", "", "set server/broker address")
	fs.IntVar(&config.Priority, prefix+"priority", 0, "set connectivity priority")
	fs.AddFlagSet(confhelper.FlagsForTLSConfig(prefix+"tls.", &config.TLS))

	return fs
}

type ArhatGRPCConfig struct {
	arhatConnectivityCommonConfig `json:",inline" yaml:",inline"`
}

func flagsForArhatGRPCConnectivityConfig(prefix string, config *ArhatGRPCConfig) *pflag.FlagSet {
	fs := pflag.NewFlagSet("arhat.conn.grpc", pflag.ExitOnError)

	fs.AddFlagSet(flagsForArhatConnectivityCommonConfig(prefix, &config.arhatConnectivityCommonConfig))

	return fs
}

type ArhatCoAPConfig struct {
	arhatConnectivityCommonConfig `json:",inline" yaml:",inline"`

	PathNamespace string            `json:"pathNamespace" yaml:"pathNamespace"`
	Transport     string            `json:"transport" yaml:"transport"`
	URIQueries    map[string]string `json:"uriQueries" yaml:"uriQueries"`
	Keepalive     int32             `json:"keepalive" yaml:"keepalive"`
}

func flagsForArhatCoAPConnectivityConfig(prefix string, config *ArhatCoAPConfig) *pflag.FlagSet {
	fs := pflag.NewFlagSet("arhat.conn.coap", pflag.ExitOnError)

	fs.AddFlagSet(flagsForArhatConnectivityCommonConfig(prefix, &config.arhatConnectivityCommonConfig))
	fs.StringVar(&config.PathNamespace, prefix+"pathNamespace", "", "set topic namespace to communicate with aranya")
	fs.StringVar(&config.Transport, prefix+"transport", "tcp",
		"set coap underlay transport protocol, one of [tcp, tcp4, tcp6, udp, udp4, udp6]")
	fs.StringToStringVar(&config.URIQueries, prefix+"uriQueries", make(map[string]string), "set coap uri queries")
	fs.Int32Var(&config.Keepalive, prefix+"keepalive", 60, "set coap keepalive interval (in seconds)")

	return fs
}

type ArhatMQTTConfig struct {
	arhatConnectivityCommonConfig `json:",inline" yaml:",inline"`

	TopicNamespace string `json:"topicNamespace" yaml:"topicNamespace"`
	Version        string `json:"version" yaml:"version"`
	Transport      string `json:"transport" yaml:"transport"`
	Username       string `json:"username" yaml:"username"`
	Password       string `json:"password" yaml:"password"`
	ClientID       string `json:"clientID" yaml:"clientID"`
	Keepalive      int32  `json:"keepalive" yaml:"keepalive"`
	Variant        string `json:"variant" yaml:"variant"`
}

func flagsForArhatMQTTConnectivityConfig(prefix string, config *ArhatMQTTConfig) *pflag.FlagSet {
	fs := pflag.NewFlagSet("arhat.conn.mqtt", pflag.ExitOnError)

	fs.AddFlagSet(flagsForArhatConnectivityCommonConfig(prefix, &config.arhatConnectivityCommonConfig))

	fs.StringVar(&config.Variant, prefix+"variant", "",
		"set mqtt implementation variant, one of [standard, azure-iot-hub, gcp-iot-core, aws-iot-core]")
	fs.StringVar(&config.TopicNamespace, prefix+"topicNamespace", "",
		"set topic namespace to communicate with aranya")
	fs.StringVar(&config.Version, prefix+"version", "3.1.1", "set mqtt version to use, one of [3.1.1]")
	fs.StringVar(&config.Transport, prefix+"transport", "tcp",
		"set mqtt underlay transport protocol, one of [tcp, websocket]")
	fs.StringVar(&config.Username, prefix+"username", "", "set mqtt username")
	fs.StringVar(&config.Password, prefix+"password", "", "set mqtt password")
	fs.StringVar(&config.ClientID, prefix+"clientid", "", "set mqtt client id")
	fs.Int32Var(&config.Keepalive, prefix+"keepalive", 60, "set mqtt keepalive interval (in seconds)")

	return fs
}

type ArhatMQTTConnectInfo struct {
	Username string
	Password string
	ClientID string

	MsgPubTopic       string
	CmdSubTopic       string
	CmdSubTopicHandle string
	WillPubTopic      string
	SupportRetain     bool

	MaxDataSize int
	TLSConfig   *tls.Config
}

func (c *ArhatMQTTConfig) GetConnectInfo() (*ArhatMQTTConnectInfo, error) {
	result := new(ArhatMQTTConnectInfo)

	variant := strings.ToLower(c.Variant)
	switch variant {
	case aranyagoconst.VariantAzureIoTHub:
		result.MaxDataSize = aranyagoconst.MaxAzureIoTHubD2CDataSize
		deviceID := c.ClientID
		result.ClientID = deviceID

		propertyBag, err := url.ParseQuery(c.TopicNamespace)
		if err != nil {
			return nil, fmt.Errorf("failed to parse property bag: %w", err)
		}
		propertyBag["dev"] = []string{deviceID}
		propertyBag["arhat"] = []string{""}
		// azure iot-hub topics
		result.MsgPubTopic = fmt.Sprintf("devices/%s/messages/events/%s", deviceID, propertyBag.Encode())
		result.WillPubTopic = result.MsgPubTopic
		result.CmdSubTopic = fmt.Sprintf("devices/%s/messages/devicebound/#", deviceID)
		result.CmdSubTopicHandle = fmt.Sprintf("devices/%s/messages/devicebound/.*", deviceID)

		result.Username = fmt.Sprintf("%s/%s/?api-version=2018-06-30", c.Endpoint, deviceID)
		// Password is set to SAS token if not using mTLS
		result.Password = c.Password
	case aranyagoconst.VariantGCPIoTCore:
		if !c.TLS.Enabled || c.TLS.Key == "" {
			return nil, fmt.Errorf("no private key found")
		}

		if c.TLS.Cert != "" {
			return nil, fmt.Errorf("cert file must be empty")
		}

		result.MaxDataSize = aranyagoconst.MaxGCPIoTCoreD2CDataSize
		result.ClientID = c.ClientID
		parts := strings.Split(c.ClientID, "/")
		if len(parts) != 8 {
			return nil, fmt.Errorf("expect 8 sections in client id but found %d", len(parts))
		}

		// second section is project id
		projectID := parts[1]
		claims := jwt.StandardClaims{
			Audience: projectID,
			IssuedAt: time.Now().Unix(),
			// valid for half a day (max value is 24 hr)
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		}

		keyBytes, err := ioutil.ReadFile(c.TLS.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key file: %w", err)
		}

		var (
			key        interface{}
			signMethod jwt.SigningMethod
		)

		block, _ := pem.Decode(keyBytes)
		switch block.Type {
		case "EC PRIVATE KEY":
			signMethod = jwt.SigningMethodES256
			key, err = x509.ParseECPrivateKey(block.Bytes)
		case "RSA PRIVATE KEY":
			signMethod = jwt.SigningMethodRS256
			key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		default:
			return nil, fmt.Errorf("unsupported private key algorithm")
		}
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		token := jwt.NewWithClaims(signMethod, claims)
		jwtToken, err := token.SignedString(key)
		if err != nil {
			return nil, fmt.Errorf("failed to sign jwt token: %w", err)
		}

		// last section is the device id
		deviceID := parts[7]
		result.MsgPubTopic = fmt.Sprintf("/devices/%s/events", deviceID)
		if c.TopicNamespace != "" {
			result.MsgPubTopic = fmt.Sprintf("/devices/%s/events/%s", deviceID, c.TopicNamespace)
		}
		result.CmdSubTopic = fmt.Sprintf("/devices/%s/commands/#", deviceID)
		result.CmdSubTopicHandle = fmt.Sprintf("/devices/%s/commands.*", deviceID)
		result.WillPubTopic = fmt.Sprintf("/devices/%s/state", deviceID)
		result.Password = jwtToken
	case aranyagoconst.VariantAWSIoTCore:
		if !c.TLS.Enabled || c.TLS.Cert == "" || c.TLS.Key == "" {
			return nil, fmt.Errorf("tls cert key pair must be provided for aws-iot-core")
		}

		result.MaxDataSize = aranyagoconst.MaxAwsIoTCoreD2CDataSize
		result.ClientID = c.ClientID
		result.CmdSubTopic, result.MsgPubTopic, result.WillPubTopic = aranyagoconst.MQTTTopics(c.TopicNamespace)
		result.CmdSubTopicHandle = result.CmdSubTopic
	case "", "standard":
		result.MaxDataSize = aranyagoconst.MaxMQTTDataSize
		result.Username = c.Username
		result.Password = c.Password
		result.ClientID = c.ClientID
		result.SupportRetain = true

		result.CmdSubTopic, result.MsgPubTopic, result.WillPubTopic = aranyagoconst.MQTTTopics(c.TopicNamespace)
		result.CmdSubTopicHandle = result.CmdSubTopic
	default:
		return nil, fmt.Errorf("unsupported variant type")
	}

	var err error
	result.TLSConfig, err = c.TLS.GetTLSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create client tls config: %w", err)
	}

	if variant == aranyagoconst.VariantAWSIoTCore {
		result.TLSConfig.NextProtos = []string{"x-amzn-mqtt-ca"}
	}

	return result, nil
}

// ArhatAppConfig configuration for arhat application behavior
type ArhatAppConfig struct {
	Log log.ConfigSet `json:"log" yaml:"log"`

	Host ArhatHostConfig `json:"host" yaml:"host"`
	Node ArhatNodeConfig `json:"node" yaml:"node"`

	Optimization struct {
		PProf         confhelper.PProfConfig `json:"pprof" yaml:"pprof"`
		MaxProcessors int                    `json:"maxProcessors" yaml:"maxProcessors"`
	} `json:"optimization" yaml:"optimization"`
}

type ArhatHostConfig struct {
	AllowAttach      bool `json:"allowAttach" yaml:"allowAttach"`
	AllowExec        bool `json:"allowExec" yaml:"allowExec"`
	AllowLog         bool `json:"allowLog" yaml:"allowLog"`
	AllowPortForward bool `json:"allowPortForward" yaml:"allowPortForward"`
}

func FlagsForArhatHostConfig(prefix string, config *ArhatHostConfig) *pflag.FlagSet {
	fs := pflag.NewFlagSet("arhat.host", pflag.ExitOnError)

	fs.BoolVar(&config.AllowAttach, prefix+"allowAttach", false, "allow kubectl attach")
	fs.BoolVar(&config.AllowExec, prefix+"allowExec", false, "allow kubectl exec")
	fs.BoolVar(&config.AllowLog, prefix+"allowLog", false, "allow kubectl logs")
	fs.BoolVar(&config.AllowPortForward, prefix+"allowPortForward", false, "allow kubectl port-forward")

	return fs
}

type ArhatNodeConfig struct {
	MachineIDFrom ArhatValueFromSpec `json:"machineIDFrom" yaml:"machineIDFrom"`
	ExtInfo       []ArhatNodeExtInfo `json:"extInfo" yaml:"extInfo"`
}

func FlagsForArhatNodeConfig(prefix string, config *ArhatNodeConfig) *pflag.FlagSet {
	fs := pflag.NewFlagSet("arhat.node", pflag.ExitOnError)

	return fs
}

type ArhatValueFromSpec struct {
	Exec []string `json:"exec" yaml:"exec"`
	File string   `json:"file" yaml:"file"`
	Text string   `json:"text" yaml:"text"`
}

func (vf *ArhatValueFromSpec) Get() (string, error) {
	if vf == nil {
		return "", nil
	}

	if vf.Text != "" {
		return vf.Text, nil
	}

	if vf.File != "" {
		data, err := ioutil.ReadFile(vf.File)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	if len(vf.Exec) > 0 {
		cmd := exechelper.Prepare(context.TODO(), vf.Exec, false, nil)
		data, err := cmd.CombinedOutput()
		if err != nil {
			return "", err
		}

		return string(data), nil
	}

	return "", nil
}

type ArhatNodeExtInfo struct {
	ValueFrom ArhatValueFromSpec `json:"valueFrom" yaml:"valueFrom"`

	ValueType string `json:"valueType" yaml:"valueType"`
	Operator  string `json:"operator" yaml:"operator"`
	ApplyTo   string `json:"applyTo" yaml:"applyTo"`
}

type ArhatRuntimeEndpoint struct {
	Endpoint      string        `json:"address" yaml:"address"`
	DialTimeout   time.Duration `json:"dialTimeout" yaml:"dialTimeout"`
	ActionTimeout time.Duration `json:"actionTimeout" yaml:"actionTimeout"`
}

func flagsForArhatRuntimeEndpointConfig(prefix string, config *ArhatRuntimeEndpoint) *pflag.FlagSet {
	fs := pflag.NewFlagSet("arhat.runtime.endpoint", pflag.ExitOnError)

	fs.StringVar(&config.Endpoint, prefix+"endpoint", "", "set endpoint address")
	fs.DurationVar(&config.DialTimeout, prefix+"dialTimeout",
		constant.DefaultRuntimeDialTimeout, "set endpoint dial timeout")
	fs.DurationVar(&config.ActionTimeout, prefix+"actionTimeout",
		constant.DefaultRuntimeActionTimeout, "set endpoint maximum action timeout")

	return fs
}

type ArhatStorageConfig struct {
	// Name of backend used to mount remote volumes
	//
	// available backend option:
	//	- "" (disabled)
	//  - sshfs
	Backend             string        `json:"backend" yaml:"backend"`
	StdoutFile          string        `json:"stdoutFile" yaml:"stdoutFile"`
	StderrFile          string        `json:"stderrFile" yaml:"stderrFile"`
	ProcessCheckTimeout time.Duration `json:"processCheckTimeout" yaml:"processCheckTimeout"`

	// LookupPaths to lookup executables required by the backend
	// will default to $PATH if not set
	LookupPaths []string `json:"lookupPaths" yaml:"lookupPaths"`

	// Args with env var references for backend mount operation
	//
	// valid env var references are
	//	 - ${ARHAT_STORAGE_REMOTE_PATH}
	//   - ${ARHAT_STORAGE_LOCAL_PATH}
	Args map[string][]string `json:"args" yaml:"args"`
}

type ArhatRuntimeConfig struct {
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
		Image ArhatRuntimeEndpoint `json:"image" yaml:"image"`
		// runtime endpoint
		Container ArhatRuntimeEndpoint `json:"container" yaml:"container"`
	} `json:"endpoints" yaml:"endpoints"`
}

func FlagsForArhatRuntimeConfig(prefix string, config *ArhatRuntimeConfig) *pflag.FlagSet {
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

func (c *ArhatRuntimeConfig) PodDir(podUID string) string {
	return filepath.Join(c.DataDir, "pods", podUID)
}

func (c *ArhatRuntimeConfig) podVolumeDir(podUID, typ, volumeName string) string {
	return filepath.Join(c.PodDir(podUID), "volumes", typ, volumeName)
}

func (c *ArhatRuntimeConfig) PodRemoteVolumeDir(podUID, volumeName string) string {
	return c.podVolumeDir(podUID, "remote", volumeName)
}

func (c *ArhatRuntimeConfig) PodBindVolumeDir(podUID, volumeName string) string {
	return c.podVolumeDir(podUID, "bind", volumeName)
}

func (c *ArhatRuntimeConfig) PodTmpfsVolumeDir(podUID, volumeName string) string {
	return c.podVolumeDir(podUID, "tmpfs", volumeName)
}

func (c *ArhatRuntimeConfig) PodResolvConfFile(podUID string) string {
	return filepath.Join(c.PodDir(podUID), "volumes", "bind", "_net", "resolv.conf")
}
