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

package storageutil

import (
	"fmt"
)

type (
	ConfigFactoryFunc func() interface{}
	FactoryFunc       func(config interface{}) (Interface, error)
)

type bundle struct {
	f  FactoryFunc
	cf ConfigFactoryFunc
}

var (
	supportedDrivers = map[string]*bundle{
		"": {
			f: func(interface{}) (Interface, error) {
				return &NopDriver{}, nil
			},
			cf: func() interface{} {
				return &NopConfig{}
			},
		},
	}
)

func Register(name string, f FactoryFunc, cf ConfigFactoryFunc) {
	if f == nil || cf == nil {
		return
	}

	// reserve empty name
	if name == "" {
		return
	}

	supportedDrivers[name] = &bundle{
		f:  f,
		cf: cf,
	}
}

func NewConfig(name string) (interface{}, error) {
	b, ok := supportedDrivers[name]
	if !ok {
		return nil, fmt.Errorf("driver %q not found", name)
	}

	return b.cf(), nil
}

func NewDriver(name string, config interface{}) (Interface, error) {
	b, ok := supportedDrivers[name]
	if !ok {
		return nil, fmt.Errorf("driver %q not found", name)
	}

	return b.f(config)
}
