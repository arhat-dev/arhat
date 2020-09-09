// +build !nosysinfo
// +build js

/*
Copyright 2019 The arhat.dev Authors.

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
	"syscall/js"

	ua "github.com/mssola/user_agent"
)

var (
	browserName, browserVersion string
)

func init() {
	defer func() {
		_ = recover()
	}()

	userAgent := ua.New(js.Global().Get("navigator").Get("userAgent").String())
	browserName, browserVersion = userAgent.Browser()
}

func GetMachineID() string {
	return ""
}

func GetTotalMemory() uint64 {
	return 0
}

func GetKernelVersion() string {
	return browserVersion
}

func GetTotalDiskSpace() uint64 {
	return 0
}

func GetOSImage() string {
	return browserName
}

func GetSystemUUID() string {
	return ""
}

func GetBootID() string {
	return ""
}
