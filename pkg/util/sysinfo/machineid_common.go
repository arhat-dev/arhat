// +build !nosysinfo
// +build !js,!plan9,!aix

package sysinfo

import "github.com/denisbrodbeck/machineid"

func GetMachineID() string {
	machineID, _ := machineid.ID()
	return machineID
}
