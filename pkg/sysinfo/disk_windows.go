// +build !nosysinfo

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
	"unsafe"

	"golang.org/x/sys/windows"
)

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
