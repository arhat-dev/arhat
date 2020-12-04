// +build noextension noextension_peripheral

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

type extensionComponentPeripheral struct{}

func (b *extensionComponentPeripheral) init(_, _, _ interface{})           {}
func (c *extensionComponentPeripheral) start(agent *Agent) error           { return nil }
func (b *extensionComponentPeripheral) RetrieveCachedMetrics() interface{} { return nil }
func (b *extensionComponentPeripheral) CacheMetrics(_ interface{})         {}
func (b *extensionComponentPeripheral) CollectMetrics(_ ...string) (_, _, _ interface{}) {
	return nil, nil, nil
}

func (b *Agent) handlePeripheralList(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "peripheral.list", nil)
}

func (b *Agent) handlePeripheralEnsure(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "peripheral.ensure", nil)
}

func (b *Agent) handlePeripheralDelete(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "peripheral.unknown", nil)
}

func (b *Agent) handlePeripheralOperate(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "peripheral.operate", nil)
}
