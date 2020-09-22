package conf

import (
	"encoding/json"
	"testing"
	"time"

	"arhat.dev/pkg/confhelper"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

type testConnectivityMethod struct {
	ConnectivityCommonConfig `json:",inline" yaml:",inline"`

	Foo string `json:"foo" yaml:"foo"`
}

func newTestConnectivityMethod(foo string) interface{} {
	return &testConnectivityMethod{
		ConnectivityCommonConfig: ConnectivityCommonConfig{
			Endpoint:       "",
			MaxPayloadSize: 0,
			TLS: confhelper.TLSConfig{
				Enabled:            true,
				CaCert:             "/path/to/ca.pem",
				Cert:               "/path/to/cert.pem",
				Key:                "/path/to/key.pem",
				ServerName:         "foo",
				InsecureSkipVerify: true,
				KeyLogFile:         "/path/to/key.log",
				CipherSuites: []string{
					"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
					"TLS_PSK_WITH_AES_128_GCM_SHA256",
				},
				PreSharedKey: confhelper.TLSPreSharedKeyConfig{
					ServerHintMapping: []string{
						"foo:bar",
					},
					IdentityHint: "foo",
				},
				AllowInsecureHashes: true,
			},
		},
		Foo: foo,
	}
}

var expectedConfig = &ConnectivityConfig{
	DialTimeout:    time.Second,
	InitialBackoff: time.Minute,
	MaxBackoff:     time.Hour,
	BackoffFactor:  1.1,
	Methods: []ConnectivityMethod{
		{
			Name:     "test",
			Priority: 1,
			Config:   newTestConnectivityMethod("1"),
		},
		{
			Name:     "test",
			Priority: 2,
			Config:   newTestConnectivityMethod("2"),
		},
		{
			Name:     "test",
			Priority: -3,
			Config:   newTestConnectivityMethod("3"),
		},
		{
			Name:     "test",
			Priority: -3,
			Config:   newTestConnectivityMethod("4"),
		},
	},
}

func TestConnectivityConfig_UnmarshalJSON(t *testing.T) {
	data, err := json.Marshal(expectedConfig)
	if !assert.NoError(t, err) {
		assert.FailNow(t, "failed to marshal expected data: %v", err)
	}

	actualConfig := new(ConnectivityConfig)
	assert.Error(t, json.Unmarshal(data, actualConfig))

	RegisterConnectivityConfig("test", func() interface{} {
		return newTestConnectivityMethod("")
	})

	actualConfig = new(ConnectivityConfig)
	assert.NoError(t, json.Unmarshal(data, actualConfig))

	_, err = json.Marshal(actualConfig)
	assert.NoError(t, err)

	for i := range expectedConfig.Methods {
		assert.EqualValues(t, expectedConfig.Methods[i].Name, actualConfig.Methods[i].Name)
		assert.EqualValues(t, expectedConfig.Methods[i].Priority, actualConfig.Methods[i].Priority)
		assert.EqualValues(t, expectedConfig.Methods[i].Config, actualConfig.Methods[i].Config)
	}
}

func TestConnectivityConfig_UnmarshalYAML(t *testing.T) {
	data, err := yaml.Marshal(expectedConfig)
	if !assert.NoError(t, err) {
		assert.FailNow(t, "failed to marshal expected data: %v", err)
	}

	actualConfig := new(ConnectivityConfig)
	assert.Error(t, yaml.UnmarshalStrict(data, actualConfig))

	RegisterConnectivityConfig("test", func() interface{} {
		return newTestConnectivityMethod("")
	})

	actualConfig = new(ConnectivityConfig)
	assert.NoError(t, yaml.Unmarshal(data, actualConfig))

	_, err = yaml.Marshal(actualConfig)
	assert.NoError(t, err)

	for i := range expectedConfig.Methods {
		assert.EqualValues(t, expectedConfig.Methods[i].Name, actualConfig.Methods[i].Name)
		assert.EqualValues(t, expectedConfig.Methods[i].Priority, actualConfig.Methods[i].Priority)
		assert.EqualValues(t, expectedConfig.Methods[i].Config, actualConfig.Methods[i].Config)
	}
}
