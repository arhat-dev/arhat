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
	"time"

	"arhat.dev/arhat/pkg/conf"

	_ "arhat.dev/arhat/pkg/client/clientadd" // add clients
)

func printErr(msg string, err error) {
	_, _ = fmt.Fprintf(os.Stderr, "%s: %v", msg, err)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	err := flags.Parse(os.Args)
	if err != nil {
		printErr("failed to parse options", err)
		os.Exit(128)
	}

	if showVersion {
		printVersion()
		return
	}

	appCtx, err = conf.ReadConfig(flags, &configFile, cliLogConfig, config)
	if err != nil {
		printErr("failed to parse config", err)
		os.Exit(128)
	}

	err = runApp(appCtx, config)
	if err != nil {
		printErr("failed to run arhat", err)
		os.Exit(1)
	}
}
