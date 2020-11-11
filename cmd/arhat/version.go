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
