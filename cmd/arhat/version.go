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

package main

import (
	"fmt"
	"os"

	"arhat.dev/arhat/pkg/version"
)

type versionOptions struct {
	branch             bool
	commit             bool
	tag                bool
	arch               bool
	goVersion          bool
	goCompilerPlatform bool
}

var (
	showVersion bool
	versionOpts = new(versionOptions)
)

func init() {
	flags.BoolVarP(&showVersion, "version", "v", false, "show all version info")
	flags.BoolVar(&versionOpts.branch, "v.branch", false, "get branch name")
	flags.BoolVar(&versionOpts.commit, "v.commit", false, "get commit hash")
	flags.BoolVar(&versionOpts.tag, "v.tag", false, "get tag name")
	flags.BoolVar(&versionOpts.arch, "v.arch", false, "get arch")
	flags.BoolVar(&versionOpts.goVersion, "v.go", false, "get go runtime/compiler version")
	flags.BoolVar(&versionOpts.goCompilerPlatform,
		"v.compilerPlatform", false, "get GOOS/GOARCH pair of the go compiler",
	)
}

func printVersion() {
	show := func(s string) {
		_, _ = fmt.Fprint(os.Stdout, s)
	}

	switch {
	case versionOpts.branch:
		show(version.Branch())
	case versionOpts.commit:
		show(version.Commit())
	case versionOpts.tag:
		show(version.Tag())
	case versionOpts.arch:
		show(version.Arch())
	case versionOpts.goVersion:
		show(version.GoVersion())
	case versionOpts.goCompilerPlatform:
		show(version.GoCompilerPlatform())
	default:
		show(version.Version())
	}
}
