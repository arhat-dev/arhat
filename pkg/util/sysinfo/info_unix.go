// +build !nosysinfo
// +build !windows,!js

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
	"io/ioutil"
	"strings"
)

func GetOSImage() string {
	osImageBytes, _ := ioutil.ReadFile("/etc/os-release")
	osImage := string(osImageBytes)

	osImageParts := strings.Split(osImage, "\n")
	for _, line := range osImageParts {
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue
		}

		if parts[0] == "PRETTY_NAME" {
			osImage = strings.Trim(parts[1], `"`)
			break
		}
	}

	return osImage
}

func GetSystemUUID() string {
	systemUUIDBytes, _ := ioutil.ReadFile("/sys/devices/virtual/dmi/id/product_uuid")

	if len(systemUUIDBytes) != 0 {
		return string(systemUUIDBytes[:len(systemUUIDBytes)-1])
	}

	return ""
}

func GetBootID() string {
	bootIDBytes, _ := ioutil.ReadFile("/proc/sys/kernel/random/boot_id")

	if len(bootIDBytes) != 0 {
		return string(bootIDBytes[:len(bootIDBytes)-1])
	}

	return ""
}
