package conf

import (
	"encoding/json"
	"fmt"
	"time"

	"arhat.dev/arhat/pkg/client"

	"gopkg.in/yaml.v2"
)

// ConnectivityConfig configuration for connectivity part in arhat
type ConnectivityConfig struct {
	DialTimeout    time.Duration `json:"dialTimeout" yaml:"dialTimeout"`
	InitialBackoff time.Duration `json:"initialBackoff" yaml:"initialBackoff"`
	MaxBackoff     time.Duration `json:"maxBackoff" yaml:"maxBackoff"`
	BackoffFactor  float64       `json:"backoffFactor" yaml:"backoffFactor"`

	Methods []ConnectivityMethod `json:"methods" yaml:"methods"`
}

type ConnectivityMethod struct {
	Name     string `json:"name" yaml:"name"`
	Priority int    `json:"priority" yaml:"priority"`

	Config interface{} `json:"config" yaml:"config"`
}

func (c *ConnectivityMethod) UnmarshalJSON(data []byte) error {
	m := make(map[string]interface{})

	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	c.Name, c.Priority, c.Config, err = unmarshalConnectivityMethod(m)
	if err != nil {
		return err
	}

	return nil
}

func (c *ConnectivityMethod) UnmarshalYAML(unmarshal func(out interface{}) error) error {
	m := make(map[string]interface{})

	err := unmarshal(&m)
	if err != nil {
		return err
	}

	c.Name, c.Priority, c.Config, err = unmarshalConnectivityMethod(m)
	if err != nil {
		return err
	}

	return nil
}

func unmarshalConnectivityMethod(m map[string]interface{}) (name string, priority int, config interface{}, err error) {
	n, ok := m["name"]
	if !ok {
		err = fmt.Errorf("must specify connectivity name")
		return
	}

	name, ok = n.(string)
	if !ok {
		err = fmt.Errorf("connectivity name must be a string")
		return
	}

	config, err = client.NewConfig(name)
	if err != nil {
		return name, 0, nil, nil
	}

	configRaw, ok := m["config"]
	if !ok {
		err = fmt.Errorf("must provide connectivity config")
		return
	}

	var configData []byte
	switch d := configRaw.(type) {
	case []byte:
		configData = d
	case string:
		configData = []byte(d)
	default:
		configData, err = yaml.Marshal(d)
		if err != nil {
			err = fmt.Errorf("failed to get connectivity config bytes: %w", err)
			return
		}
	}

	err = yaml.UnmarshalStrict(configData, config)
	if err != nil {
		return
	}

	p := m["priority"]
	switch d := p.(type) {
	case int:
		priority = d
	case int8:
		priority = int(d)
	case int16:
		priority = int(d)
	case int32:
		priority = int(d)
	case int64:
		priority = int(d)
	case uint:
		priority = int(d)
	case uint8:
		priority = int(d)
	case uint16:
		priority = int(d)
	case uint32:
		priority = int(d)
	case uint64:
		priority = int(d)
	case float32:
		priority = int(d)
	case float64:
		priority = int(d)
	default:
		err = fmt.Errorf("unexpected priority type, must be an integer")
		return
	}

	return name, priority, config, nil
}
