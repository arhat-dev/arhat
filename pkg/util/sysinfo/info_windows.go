// +build !nosysinfo
// +build windows

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

package sysinfo

import (
	"fmt"
	"os/exec"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	kernel32DLL         *windows.DLL
	getDiskFreeSpaceExW *windows.Proc
)

var (
	osImage       string
	kernelVersion string
)

func init() {
	kernel32DLL = windows.MustLoadDLL("kernel32.dll")
	getDiskFreeSpaceExW = kernel32DLL.MustFindProc("GetDiskFreeSpaceExW")

	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return
	}
	defer func() {
		_ = k.Close()
	}()

	{
		// get product name as os image
		osImage, _, _ = k.GetStringValue("ProductName")
	}

	{
		// build kernel version
		buildNumber, _, err := k.GetStringValue("CurrentBuildNumber")
		if err != nil {
			return
		}

		majorVersionNumber, _, err := k.GetIntegerValue("CurrentMajorVersionNumber")
		if err != nil {
			return
		}

		minorVersionNumber, _, err := k.GetIntegerValue("CurrentMinorVersionNumber")
		if err != nil {
			return
		}

		revision, _, err := k.GetIntegerValue("UBR")
		if err != nil {
			return
		}

		kernelVersion = fmt.Sprintf("%d.%d.%s.%d", majorVersionNumber, minorVersionNumber, buildNumber, revision)
	}
}

func GetKernelVersion() string {
	return kernelVersion
}

func GetTotalDiskSpace() uint64 {
	_, totalBytes := checkDisk("")
	return totalBytes
}

func checkDisk(dataDir string) (freeBytes, totalBytes uint64) {
	var totalFreeBytes uint64

	ptr, err := windows.UTF16PtrFromString(dataDir)
	if err != nil {
		return
	}

	_, _, _ = getDiskFreeSpaceExW.Call(uintptr(unsafe.Pointer(ptr)),
		uintptr(unsafe.Pointer(&freeBytes)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)

	return
}

func getFreeDiskSpace() uint64 {
	freeBytes, _ := checkDisk("")
	return freeBytes
}

func GetOSImage() string { return osImage }
func GetBootID() string  { return "" }

func GetSystemUUID() string {
	result, err := exec.Command("wmic", "csproduct", "get", "UUID").Output()
	if err != nil {
		return ""
	}

	fields := strings.Fields(string(result))
	if len(fields) != 2 {
		return ""
	}

	return fields[1]
}
