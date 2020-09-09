// +build nosysinfo

package sysinfo

func GetMachineID() string      { return "" }
func GetTotalMemory() uint64    { return 0 }
func GetKernelVersion() string  { return "" }
func GetTotalDiskSpace() uint64 { return 0 }
func GetOSImage() string        { return "" }
func GetSystemUUID() string     { return "" }
func GetBootID() string         { return "" }
