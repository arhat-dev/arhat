// +build !nonethelper
// +build !nonethelper_piondtls

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

package piondtls

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/pion/dtls/v2"
	"github.com/pion/udp"

	"arhat.dev/pkg/nethelper"
)

func init() {
	for _, n := range []string{"udp", "udp4", "udp6"} {
		nethelper.Register(n, true, listen, dial)
	}
}

func dial(
	ctx context.Context,
	dialer interface{},
	network, addr string,
	tlsCfg interface{},
) (net.Conn, error) {
	var tlsConfig *dtls.Config
	switch t := tlsCfg.(type) {
	case *tls.Config:
		tlsConfig = convertTLSConfig(t)
	case *dtls.Config:
		tlsConfig = t
	default:
		return nil, nethelper.ErrTLSConfigInvalid
	}

	var (
		conn net.Conn
		err  error
	)
	if dialer != nil {
		d, ok := dialer.(*net.Dialer)
		if !ok {
			return nil, nethelper.ErrDialerInvalid
		}

		conn, err = d.DialContext(ctx, network, addr)
	} else {
		conn, err = (&net.Dialer{}).DialContext(ctx, network, addr)
	}
	if err != nil {
		return nil, err
	}

	return dtls.ClientWithContext(ctx, conn, tlsConfig)
}

func listen(
	ctx context.Context,
	config interface{},
	network, addr string,
	tlsCfg interface{},
) (interface{}, error) {
	_ = ctx

	var tlsConfig *dtls.Config
	switch t := tlsCfg.(type) {
	case *tls.Config:
		tlsConfig = convertTLSConfig(t)
	case *dtls.Config:
		tlsConfig = t
	default:
		return nil, nethelper.ErrTLSConfigInvalid
	}

	var lc *udp.ListenConfig
	switch t := config.(type) {
	case *udp.ListenConfig:
		lc = t
	case *net.ListenConfig:
		// unused
	case nil:
		// allow empty config
	default:
		return nil, nethelper.ErrListenConfigInvalid
	}

	laddr, err := net.ResolveUDPAddr(network, addr)
	if err != nil {
		return nil, err
	}

	if lc == nil {
		return dtls.Listen(network, laddr, tlsConfig)
	}

	l, err := lc.Listen(network, laddr)
	if err != nil {
		return nil, err
	}
	return dtls.NewListener(l, tlsConfig)
}

func convertTLSConfig(tlsConfig *tls.Config) *dtls.Config {
	if tlsConfig == nil {
		return nil
	}

	var cs []dtls.CipherSuiteID
	for i := range tlsConfig.CipherSuites {
		cs = append(cs, dtls.CipherSuiteID(tlsConfig.CipherSuites[i]))
	}

	return &dtls.Config{
		Certificates:          tlsConfig.Certificates,
		CipherSuites:          cs,
		InsecureSkipVerify:    tlsConfig.InsecureSkipVerify,
		VerifyPeerCertificate: tlsConfig.VerifyPeerCertificate,
		RootCAs:               tlsConfig.RootCAs,
		ClientCAs:             tlsConfig.ClientCAs,
		ServerName:            tlsConfig.ServerName,
		ClientAuth:            dtls.ClientAuthType(tlsConfig.ClientAuth),

		// TODO: support more dTLS options
		SignatureSchemes:       nil,
		SRTPProtectionProfiles: nil,
		ExtendedMasterSecret:   0,
		FlightInterval:         0,
		PSK:                    nil,
		PSKIdentityHint:        nil,
		InsecureHashes:         false,
		LoggerFactory:          nil,
		MTU:                    0,
		ReplayProtectionWindow: 0,
	}
}
