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

package util

import (
	"crypto/tls"

	"github.com/pion/dtls/v2"
)

func ConvertTLSConfigToDTLSConfig(tlsConfig *tls.Config) *dtls.Config {
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

		// TODO: support more dTLS options
		SignatureSchemes:       nil,
		SRTPProtectionProfiles: nil,
		ClientAuth:             0,
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
