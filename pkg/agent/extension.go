// +build !noextension

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

package agent

import (
	"crypto/tls"
	"fmt"

	"arhat.dev/libext/server"
	"arhat.dev/pkg/log"

	"arhat.dev/arhat/pkg/conf"
)

type agentComponentExtension struct {
	srv *server.Server
	extensionComponentPeripheral
	extensionComponentRuntime
}

func (c *agentComponentExtension) init(
	agent *Agent,
	logger log.Interface,
	config *conf.ExtensionConfig,
) error {
	if !config.Enabled {
		return nil
	}

	var endpoints []server.EndpointConfig
	for _, ep := range config.Endpoints {
		tlsConfig, err := ep.TLS.TLSConfig.GetTLSConfig(true)
		if err != nil {
			return fmt.Errorf("failed to create tls config for extension endpoint %q: %w", ep.Listen, err)
		}

		if tlsConfig != nil && ep.TLS.VerifyClientCert {
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}

		endpoints = append(endpoints, server.EndpointConfig{
			Listen:            ep.Listen,
			TLS:               tlsConfig,
			KeepaliveInterval: ep.KeepaliveInterval,
			MessageTimeout:    ep.MessageTimeout,
		})
	}

	var err error
	c.srv, err = server.NewServer(agent.ctx, logger, &server.Config{
		Endpoints: endpoints,
	})
	if err != nil {
		return fmt.Errorf("failed to create extension server: %w", err)
	}

	c.extensionComponentPeripheral.init(agent, c.srv, &config.Peripheral)
	c.extensionComponentRuntime.init(agent, c.srv, &config.Runtime)

	go func() {
		err2 := c.srv.ListenAndServe()
		if err2 != nil {
			panic(err2)
		}
	}()

	err = c.extensionComponentPeripheral.start(agent)
	if err != nil {
		return fmt.Errorf("failed to start peripheral manager: %w", err)
	}

	err = c.extensionComponentRuntime.start(agent)
	if err != nil {
		return fmt.Errorf("failed to start runtime manager: %w", err)
	}

	return nil
}
