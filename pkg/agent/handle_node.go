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
	"runtime"

	"arhat.dev/aranya-proto/aranyagopb"

	"arhat.dev/arhat/pkg/sysinfo"
	"arhat.dev/arhat/pkg/version"
)

func (b *Agent) handleNodeInfoGet(sid uint64, cmdBytes []byte) {
	cmd := new(aranyagopb.NodeInfoGetCmd)
	err := cmd.Unmarshal(cmdBytes)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal NodeInfoGetCmd: %w", err))
		return
	}

	switch cmd.Kind {
	case aranyagopb.NODE_INFO_ALL:
		b.processInNewGoroutine(sid, "node.info.all", func() {
			capacity := &aranyagopb.NodeResources{
				CpuCount:     uint64(runtime.NumCPU()),
				MemoryBytes:  sysinfo.GetTotalMemory(),
				StorageBytes: sysinfo.GetTotalDiskSpace(),
			}

			machineID, _ := b.machineIDFrom.Get()
			if machineID == "" {
				machineID = sysinfo.GetMachineID()
			}

			systemInfo := &aranyagopb.NodeSystemInfo{
				Os:            runtime.GOOS,
				Arch:          version.Arch(),
				OsImage:       sysinfo.GetOSImage(),
				KernelVersion: sysinfo.GetKernelVersion(),
				MachineId:     machineID,
				BootId:        sysinfo.GetBootID(),
				SystemUuid:    sysinfo.GetSystemUUID(),
			}

			nodeMsg := &aranyagopb.NodeStatusMsg{
				SystemInfo: systemInfo,
				Capacity:   capacity,
				Conditions: b.getNodeConditions(),
				ExtInfo:    b.extInfo,
			}
			if err := b.PostMsg(sid, aranyagopb.MSG_NODE_STATUS, nodeMsg); err != nil {
				b.handleConnectivityError(sid, err)
				return
			}
		})
	case aranyagopb.NODE_INFO_DYN:
		b.processInNewGoroutine(sid, "node.info.dyn", func() {
			nodeMsg := &aranyagopb.NodeStatusMsg{
				Conditions: b.getNodeConditions(),
			}
			if err := b.PostMsg(sid, aranyagopb.MSG_NODE_STATUS, nodeMsg); err != nil {
				b.handleConnectivityError(sid, err)
				return
			}
		})
	default:
		b.handleUnknownCmd(sid, "node", cmd)
	}
}

func (b *Agent) getNodeConditions() *aranyagopb.NodeConditions {
	// TODO: use real conditions
	return &aranyagopb.NodeConditions{
		Ready:   aranyagopb.NODE_CONDITION_HEALTHY,
		Memory:  aranyagopb.NODE_CONDITION_HEALTHY,
		Disk:    aranyagopb.NODE_CONDITION_HEALTHY,
		Pid:     aranyagopb.NODE_CONDITION_HEALTHY,
		Network: aranyagopb.NODE_CONDITION_HEALTHY,
		Pod:     aranyagopb.NODE_CONDITION_HEALTHY,
	}
}
