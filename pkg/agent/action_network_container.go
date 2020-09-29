// +build !rt_none

package agent

import (
	"fmt"

	"arhat.dev/aranya-proto/aranyagopb"

	"arhat.dev/arhat/pkg/types"
)

func (b *Agent) handleContainerNetworkList(sid uint64, data []byte) {
	cmd := new(aranyagopb.ContainerNetworkListCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal ContainerNetworkListCmd: %w", err))
		return
	}

	rcn, ok := b.runtime.(types.RuntimeContainerNetwork)
	if !ok {
		b.handleUnknownCmd(sid, "net.ctr.list", cmd)
		return
	}

	b.processInNewGoroutine(sid, "net.ctr.list", func() {
		result, err := rcn.GetContainerNetworkConfig()
		if err != nil {
			b.handleRuntimeError(sid, err)
		}

		err = b.PostMsg(sid, aranyagopb.MSG_CTR_NET_STATUS, result)
		if err != nil {
			b.handleConnectivityError(sid, err)
		}
	})
}

func (b *Agent) handleContainerNetworkEnsure(sid uint64, data []byte) {
	cmd := new(aranyagopb.ContainerNetworkEnsureCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal ContainerNetworkEnsureCmd: %w", err))
		return
	}

	rcn, ok := b.runtime.(types.RuntimeContainerNetwork)
	if !ok {
		b.handleUnknownCmd(sid, "net.ctr.ensure", cmd)
		return
	}

	b.processInNewGoroutine(sid, "net.ctr.ensure", func() {
		result, err := rcn.EnsureContainerNetwork(cmd)
		if err != nil {
			b.handleRuntimeError(sid, err)
		}

		msg := aranyagopb.NewPodStatusListMsg(result)
		err = b.PostMsg(sid, aranyagopb.MSG_POD_STATUS_LIST, msg)
		if err != nil {
			b.handleConnectivityError(sid, err)
		}
	})
}
