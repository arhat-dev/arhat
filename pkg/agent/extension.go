// +build !noextension

package agent

import (
	"crypto/tls"
	"fmt"

	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/libext/server"
	"arhat.dev/pkg/log"
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

	return nil
}
