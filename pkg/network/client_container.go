// +build !rt_none

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

package network

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math"
	"strings"

	"arhat.dev/abbot-proto/abbotgopb"
	"arhat.dev/aranya-proto/aranyagopb"
)

// EnsureContainerNetwork will ensure container's network meets requirements in options
func (c *Client) EnsureContainerNetwork(options *aranyagopb.ContainerNetworkEnsureCmd) error {
	var (
		ipv4Subnet = options.Ipv4Cidr
		ipv6Subnet = options.Ipv6Cidr
	)

	if ipv4Subnet == "" && ipv6Subnet == "" {
		return fmt.Errorf("no ipv4 or ipv6 subnet provided")
	}

	_, err := c.doRequest(newReqForConfigUpdate(ipv4Subnet, ipv6Subnet))
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) EnsurePodNetwork(
	namespace, name string,
	ctrID string,
	pid uint32,
	opts *aranyagopb.PodNetworkSpec,
) (ipv4, ipv6 string, err error) {
	var capArgs []*abbotgopb.CNICapArgs
	capArgs = append(capArgs, &abbotgopb.CNICapArgs{
		Option: &abbotgopb.CNICapArgs_DnsConfigArg{
			DnsConfigArg: &abbotgopb.CNICapArgs_DNSConfig{
				Servers:  opts.NameServers,
				Searches: opts.SearchDomains,
				Options:  opts.DnsOptions,
			},
		},
	})

	// ipRange cap arg for cni
	if opts.CidrIpv4 != "" {
		capArgs = append(capArgs, &abbotgopb.CNICapArgs{
			Option: &abbotgopb.CNICapArgs_IpRangeArg{
				IpRangeArg: &abbotgopb.CNICapArgs_IPRange{
					Subnet: opts.CidrIpv4,
				},
			},
		})
	}

	if opts.CidrIpv6 != "" {
		capArgs = append(capArgs, &abbotgopb.CNICapArgs{
			Option: &abbotgopb.CNICapArgs_IpRangeArg{
				IpRangeArg: &abbotgopb.CNICapArgs_IPRange{
					Subnet: opts.CidrIpv6,
				},
			},
		})
	}

	if b := opts.Bandwidth; b != nil {
		capArgs = append(capArgs, &abbotgopb.CNICapArgs{
			Option: &abbotgopb.CNICapArgs_BandwidthArg{
				BandwidthArg: &abbotgopb.CNICapArgs_Bandwidth{
					IngressRate: b.IngressRate,
					EgressRate:  b.EgressRate,
					// currently it's unlimited in kubelet
					IngressBurst: math.MaxInt32,
					EgressBurst:  math.MaxInt32,
				},
			},
		})
	}

	for _, port := range opts.Ports {
		// portMapping cap args for cni
		capArgs = append(capArgs, &abbotgopb.CNICapArgs{
			Option: &abbotgopb.CNICapArgs_PortMapArg{
				PortMapArg: &abbotgopb.CNICapArgs_PortMap{
					ContainerPort: port.ContainerPort,
					HostPort:      port.HostPort,
					Protocol:      strings.ToLower(port.Protocol),
					HostIp:        port.HostIp,
				},
			},
		})
	}

	result, err := c.doRequest(newReqForLinkEnsure(namespace, name, ctrID, pid, capArgs))
	if err != nil {
		return "", "", err
	}

	// TODO: support ipv6
	_ = result
	return
}

func (c *Client) DeletePodNetwork(ctrID string, pid uint32) error {
	_, err := c.doRequest(newReqForLinkDelete(ctrID, pid))
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) RestorePodNetwork(ctrID string, pid uint32) error {
	_, err := c.doRequest(newReqForRestoreAddress(ctrID, pid))
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetPodIPAddresses(pid uint32) (ipv4, ipv6 string, err error) {
	result, err := c.doRequest(newReqForGetAddress(pid))
	if err != nil {
		return "", "", err
	}

	_ = result
	// TODO: Update abbot-proto
	return
}

func encodeRequest(req *abbotgopb.Request) (string, error) {
	data, err := req.Marshal()
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

// nolint:unparam
func (c *Client) doRequest(req *abbotgopb.Request, err error) (*abbotgopb.Response, error) {
	if err != nil {
		return nil, err
	}

	encodedReq, err := encodeRequest(req)
	if err != nil {
		return nil, err
	}

	output := new(bytes.Buffer)
	err = c.execAbbot([]string{"request", encodedReq}, output)
	result := output.String()

	if err != nil {
		if result != "" {
			return nil, fmt.Errorf("%s: %w", result, err)
		}

		return nil, err
	}

	respBytes, err := base64.StdEncoding.DecodeString(result)
	if err != nil {
		// not base64 encoded, error happened
		return nil, fmt.Errorf(result)
	}

	resp := new(abbotgopb.Response)
	err = resp.Unmarshal(respBytes)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func newReqForLinkEnsure(
	podNamespace, podName string,
	pauseCtrID string,
	pid uint32,
	capArgs []*abbotgopb.CNICapArgs,
) (*abbotgopb.Request, error) {
	cniArgs := map[string]string{
		"IgnoreUnknown":              "true",
		"K8S_POD_NAMESPACE":          podNamespace,
		"K8S_POD_NAME":               podName,
		"K8S_POD_INFRA_CONTAINER_ID": pauseCtrID,
	}

	// TODO: support static `IP` CNI_ARG

	return abbotgopb.NewRequest(abbotgopb.NewContainerNetworkEnsureRequest(pauseCtrID, pid, capArgs, cniArgs))
}

func newReqForLinkDelete(pauseCtrID string, pid uint32) (*abbotgopb.Request, error) {
	return abbotgopb.NewRequest(abbotgopb.NewContainerNetworkDeleteRequest(pauseCtrID, pid))
}

func newReqForConfigUpdate(ipv4Subnet, ipv6Subnet string) (*abbotgopb.Request, error) {
	return abbotgopb.NewRequest(abbotgopb.NewContainerNetworkConfigEnsureRequest(ipv4Subnet, ipv6Subnet))
}

func newReqForGetAddress(pid uint32) (*abbotgopb.Request, error) {
	return abbotgopb.NewRequest(abbotgopb.NewContainerNetworkQueryRequest("", pid))
}

func newReqForRestoreAddress(containerID string, pid uint32) (*abbotgopb.Request, error) {
	return abbotgopb.NewRequest(abbotgopb.NewContainerNetworkRestoreRequest(containerID, pid))
}
