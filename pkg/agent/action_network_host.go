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
	"fmt"
	"net"

	"arhat.dev/aranya-proto/aranyagopb"
)

func (b *Agent) handleHostNetworkList(sid uint64, data []byte) {
	cmd := new(aranyagopb.HostNetworkListCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal HostNetworkListCmd: %w", err))
		return
	}

	b.processInNewGoroutine(sid, "net.host.list", func() {
		var interfaces []*aranyagopb.HostNetworkInterface
		if len(cmd.InterfaceNames) == 0 {
			var ifaces []net.Interface
			ifaces, err = net.Interfaces()
			for i := range ifaces {
				var iface *aranyagopb.HostNetworkInterface
				iface, err = generateInterface(&ifaces[i])
				if err != nil {
					break
				}

				interfaces = append(interfaces, iface)
			}
		} else {
			for _, n := range cmd.InterfaceNames {
				var iface *net.Interface
				iface, err = net.InterfaceByName(n)
				if err != nil {
					break
				}

				var f *aranyagopb.HostNetworkInterface
				f, err = generateInterface(iface)
				if err != nil {
					break
				}

				interfaces = append(interfaces, f)
			}
		}
		if err != nil {
			b.handleRuntimeError(sid, err)
			return
		}

		err = b.PostMsg(sid, aranyagopb.MSG_HOST_NET_STATUS, aranyagopb.NewHostNetworkStatusMsg(interfaces))
		if err != nil {
			b.handleConnectivityError(sid, err)
		}
	})
}

func generateInterface(iface *net.Interface) (*aranyagopb.HostNetworkInterface, error) {
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	var addresses []string
	for _, addr := range addrs {
		addresses = append(addresses, addr.String())
	}

	return &aranyagopb.HostNetworkInterface{
		Name:            iface.Name,
		HardwareAddress: iface.HardwareAddr.String(),
		IpAddresses:     addresses,
	}, nil
}
