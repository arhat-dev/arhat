package types

import "io"

type DelegateExecFunc func(cmd []string, stdout, stderr io.Writer) error

type NetworkClient interface {
	CreateResolvConf(nameservers, searches, options []string) ([]byte, error)

	// DelegateExec delegates abbot proto requests as command execution
	// pid and container id are optional and only used for container network
	DelegateExec(abbotReqData []byte, pid int64, containerID string) (abbotRespData []byte, err error)

	// RestoreContainerNetwork host network interfaces or container networks
	// if not pid or containerID specified, restore host network
	RestoreContainerNetwork(pid int64, containerID string) error

	// QueryContainerNetwork
	QueryContainerNetwork(pid int64, containerID string) ([]byte, error)

	// DeleteContainerNetwork
	DeleteContainerNetwork(pid int64, containerID string) error
}
