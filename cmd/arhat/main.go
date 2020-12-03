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
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/arhat/pkg/exec"

	// connectivity methods
	_ "arhat.dev/arhat/pkg/client/coap" // add coap client support
	_ "arhat.dev/arhat/pkg/client/grpc" // add grpc client support
	_ "arhat.dev/arhat/pkg/client/mqtt" // add mqtt client support
	_ "arhat.dev/arhat/pkg/client/nats" // add nats streaming client support

	// extension and port-forward network support
	_ "arhat.dev/pkg/nethelper/piondtls" // add udp dtls network support
	_ "arhat.dev/pkg/nethelper/pipenet"  // add pipe network support
	_ "arhat.dev/pkg/nethelper/stdnet"   // add standard library network support

	// extension codec support
	_ "arhat.dev/libext/codec/gogoprotobuf" // Add protobuf codec support.
	_ "arhat.dev/libext/codec/stdjson"      // Add json codec support.

	// storage drivers
	_ "ext.arhat.dev/runtimeutil/storageutil/general" // add general purpose storage driver
	_ "ext.arhat.dev/runtimeutil/storageutil/sshfs"   // add sshfs storage driver
)

func printErr(msg string, err error) {
	_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	bin := filepath.Base(os.Args[0])
	err := flags.Parse(os.Args)
	if err != nil {
		// not valid args, can be symlinked targets
		startedCmd, err2 := exec.DoIfTryFailed(
			os.Stdin, os.Stdout, os.Stderr,
			append([]string{bin}, os.Args[1:]...),
			false, nil, true,
		)
		if err2 != nil {
			printErr("failed to run as command "+bin, err2)
			printErr("failed to parse options", err)
			os.Exit(128)
		}

		exitCode, err2 := startedCmd.Wait()
		if err2 != nil {
			printErr("failed to run as command "+bin, err2)
		}
		os.Exit(exitCode)
	}

	if showVersion {
		printVersion()
		return
	}

	appCtx, err = conf.ReadConfig(flags, &configFile, cliLogConfig, config)
	if err != nil {
		// not valid args, can be symlinked targets
		startedCmd, err2 := exec.DoIfTryFailed(
			os.Stdin, os.Stdout, os.Stderr,
			append([]string{bin}, os.Args[1:]...),
			false, nil, true,
		)
		if err2 != nil {
			printErr("failed to run as command "+bin, err2)
			printErr("failed to parse config", err)
			os.Exit(128)
		}

		exitCode, err2 := startedCmd.Wait()
		if err2 != nil {
			printErr("failed to run as command "+bin, err2)
		}
		os.Exit(exitCode)
	}

	err = runApp(appCtx, config)
	if err != nil {
		printErr("failed to run arhat", err)
		os.Exit(1)
	}
}
