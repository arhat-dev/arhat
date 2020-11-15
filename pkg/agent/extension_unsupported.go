// +build noextension

package agent

type agentComponentExtension struct {
	extensionComponentPeripheral
	extensionComponentRuntime
}

func (c *agentComponentExtension) init(_, _, _ interface{}) error { return nil }
