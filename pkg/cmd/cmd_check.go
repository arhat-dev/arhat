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

package cmd

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"unsafe"

	"github.com/gogo/protobuf/proto"

	"arhat.dev/aranya-proto/aranyagopb/aranyagoconst"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"arhat.dev/aranya-proto/aranyagopb"

	"arhat.dev/arhat/pkg/client"
	"arhat.dev/arhat/pkg/client/impl"
	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/arhat/pkg/constant"
	"arhat.dev/arhat/pkg/runtime"
	"arhat.dev/arhat/pkg/runtime/none"
	"arhat.dev/arhat/pkg/storage"
	"arhat.dev/arhat/pkg/types"
)

type checkOpts struct {
	log          bool
	metrics      bool
	runtime      bool
	connectivity bool
	storage      bool
}

const checkCount = unsafe.Sizeof(checkOpts{}) / unsafe.Sizeof(true)

func newCheckCmd(appCtx *context.Context) *cobra.Command {
	var (
		opts     = new(checkOpts)
		checkAll bool
	)

	checkCmd := &cobra.Command{
		Use:           "check",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch {
			case opts.log, opts.metrics, opts.runtime, opts.connectivity, opts.storage:
				checkAll = false
			}

			if checkAll {
				opts.log = true
				opts.metrics = true
				opts.runtime = true
				opts.connectivity = true
				opts.storage = true
			}

			return runCheck(*appCtx, (*appCtx).Value(constant.ContextKeyConfig).(*conf.ArhatConfig), opts)
		},
	}

	flags := checkCmd.Flags()

	flags.BoolVar(&checkAll, "all", true, "check all")
	flags.BoolVar(&opts.log, "log", false, "check log config")
	flags.BoolVar(&opts.metrics, "metrics", false, "check metrics collectors")
	flags.BoolVar(&opts.runtime, "runtime", false, "check runtime")
	flags.BoolVar(&opts.connectivity, "connectivity", false, "check connectivity")
	flags.BoolVar(&opts.storage, "storage", false, "check storage")

	return checkCmd
}

func runCheck(appCtx context.Context, config *conf.ArhatConfig, opts *checkOpts) error {
	showResult := func(name string, val interface{}) {
		if err, ok := val.(error); ok {
			fmt.Printf("%s: err(%v)\n", name, err)
		} else {
			fmt.Printf("%s: %v\n", name, val)
		}
	}

	configBytes, err := yaml.Marshal(config)
	if err != nil {
		showResult("config", err)
	} else {
		fmt.Println("config: |")
		s := bufio.NewScanner(bytes.NewReader(configBytes))
		s.Split(bufio.ScanLines)
		for s.Scan() {
			fmt.Printf("  %s\n", s.Text())
		}
	}

	fmt.Println("---")

	for i := checkCount; i > 0; i-- {
		switch {
		case opts.log:
			opts.log = false
			for i, logConfig := range config.Arhat.Log.GetUnique() {
				showResult(fmt.Sprintf("log[%d]", i),
					fmt.Sprintf("%s@%s:%s", logConfig.Destination.File, logConfig.Format, logConfig.Level))
			}
			showResult("log.kube", config.Arhat.Log.KubeLogFile())
		case opts.runtime:
			opts.runtime = false
			var (
				rt  types.Runtime
				err error
			)

			if config.Runtime.Enabled {
				rt, err = runtime.NewRuntime(appCtx, nil, &config.Runtime)
			} else {
				rt, err = none.NewNoneRuntime(appCtx, nil, nil)
			}

			if err != nil {
				showResult("runtime", err)
			} else {
				showResult("runtime", fmt.Sprintf("%s://%s", rt.Name(), rt.Version()))
			}
		case opts.storage:
			opts.storage = false
			s, err := storage.NewStorage(appCtx, &config.Storage)
			if err != nil {
				showResult("storage", err)
				continue
			}

			showResult("storage", s.Name())
		case opts.connectivity:
			opts.connectivity = false
			c, err := client.NewClient(new(fakeAgent), &config.Connectivity.ArhatConnectivityMethods)
			if err != nil {
				showResult("connectivity", err)
				continue
			}

			method := "unknown"
			switch c.(type) {
			case *impl.GRPCClient:
				method = aranyagoconst.ConnectivityMethodGRPC
			case *impl.MQTTClient:
				method = aranyagoconst.ConnectivityMethodMQTT
			case *impl.CoAPClient:
				method = aranyagoconst.ConnectivityMethodCoAP
			}
			showResult("connectivity.method", method)

			func() {
				defer func() { _ = c.Close() }()
				err := c.Connect(context.TODO())
				if err != nil {
					showResult("connectivity.conn", err)
				} else {
					showResult("connectivity.conn", "ok")
				}
			}()
		}
	}

	return nil
}

type fakeAgent struct{}

func (a *fakeAgent) Context() context.Context      { return context.TODO() }
func (a *fakeAgent) HandleCmd(cmd *aranyagopb.Cmd) {}
func (a *fakeAgent) PostMsg(sid uint64, kind aranyagopb.Kind, msg proto.Marshaler) error {
	return nil
}
func (a *fakeAgent) PostData(sid uint64, kind aranyagopb.Kind, seq uint64, completed bool, data []byte) (lastSeq uint64, _ error) {
	return 0, nil
}
