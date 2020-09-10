// +build !rt_libpod

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
	"strings"
	"time"

	"github.com/spf13/pflag"

	"arhat.dev/arhat/pkg/cmd"
	"arhat.dev/arhat/pkg/version"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	rootCmd := cmd.NewArhatCmd()
	rootCmd.AddCommand(version.NewVersionCmd())

	rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		switch {
		case strings.HasPrefix(f.Name, "runtime"), strings.HasPrefix(f.Name, "pod"):
			f.Hidden = true
		}
	})

	err := rootCmd.Execute()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to run arhat %v: %v\n", os.Args, err)
		os.Exit(1)
	}
}
