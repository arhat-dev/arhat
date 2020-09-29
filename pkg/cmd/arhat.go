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
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sort"
	"time"

	"arhat.dev/pkg/backoff"
	"arhat.dev/pkg/confhelper"
	"arhat.dev/pkg/log"
	"github.com/spf13/cobra"

	"arhat.dev/arhat/pkg/agent"
	"arhat.dev/arhat/pkg/client"
	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/arhat/pkg/constant"
	"arhat.dev/arhat/pkg/util/manager"

	_ "arhat.dev/arhat/pkg/client/clientadd" // add clients
)

func NewArhatCmd() *cobra.Command {
	var (
		appCtx       context.Context
		configFile   string
		config       = new(conf.ArhatConfig)
		cliLogConfig = new(log.Config)
	)

	arhatCmd := &cobra.Command{
		Use:           "arhat",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Use == "version" {
				return nil
			}

			var err error
			appCtx, err = conf.ReadConfig(cmd, &configFile, cliLogConfig, config)
			if err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(appCtx, config)
		},
	}

	flags := arhatCmd.PersistentFlags()
	// config file
	flags.StringVarP(&configFile, "config", "c", constant.DefaultArhatConfigFile, "path to the arhat config file")
	// agent flags
	flags.AddFlagSet(log.FlagsForLogConfig("log.", cliLogConfig))
	// agent host options
	flags.AddFlagSet(conf.FlagsForArhatHostConfig("host.", &config.Arhat.Host))
	// optimization options
	flags.AddFlagSet(confhelper.FlagsForPProfConfig("opt.pprof.", &config.Arhat.Optimization.PProf))
	flags.IntVar(&config.Arhat.Optimization.MaxProcessors, "opt.maxProcessors", 0, "set GOMAXPROCS")
	// runtime flags
	flags.AddFlagSet(conf.FlagsForArhatRuntimeConfig("runtime.", &config.Runtime))

	// storage flags
	flags.DurationVar(
		&config.Storage.ProcessCheckTimeout,
		"storage.processCheckTimeout",
		5*time.Second,
		"assume command execution successful after this time period",
	)

	flags.AddFlagSet(conf.FlagsForExtensionConfig("ext.", &config.Extension))

	arhatCmd.AddCommand(newCheckCmd(&appCtx))

	return arhatCmd
}

func run(appCtx context.Context, config *conf.ArhatConfig) error {
	runtime.GOMAXPROCS(config.Arhat.Optimization.MaxProcessors)

	logger := log.Log.WithName("cmd")

	// handle pprof
	if cfg := config.Arhat.Optimization.PProf; cfg.Enabled {
		mgr, err := manager.NewPProfManager(cfg.Listen, cfg.HTTPPath, cfg.MutexProfileFraction, cfg.BlockProfileRate)
		if err != nil {
			return fmt.Errorf("failed to listen tcp for pprof server: %w", err)
		}

		logger.D("serving pprof stats")
		go func() {
			if err := mgr.Start(); err != nil && err != http.ErrServerClosed {
				logger.E("failed to start pprof server", log.Error(err))
				os.Exit(1)
			}
		}()
	}

	ag, err := agent.NewAgent(appCtx, config)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	exiting := func() bool {
		select {
		case <-appCtx.Done():
			return true
		case <-ag.Context().Done():
			return true
		default:
			return false
		}
	}

	bs := backoff.NewStrategy(
		config.Connectivity.InitialBackoff,
		config.Connectivity.MaxBackoff,
		config.Connectivity.BackoffFactor,
		1,
	)
	backoffTimer := time.NewTimer(0)
	if !backoffTimer.Stop() {
		<-backoffTimer.C
	}
	defer backoffTimer.Stop()

	sort.SliceStable(config.Connectivity.Methods, func(i, j int) bool {
		return config.Connectivity.Methods[i].Priority < config.Connectivity.Methods[j].Priority
	})

	for !exiting() {
		for i := 0; i < len(config.Connectivity.Methods) && !exiting(); i++ {
			name := config.Connectivity.Methods[i].Name
			id := fmt.Sprintf("%s@%d", name, i)

			logger.I("creating client", log.String("id", id))
			cl, err := client.NewConnectivityClient(
				name, ag, config.Connectivity.Methods[i].Config,
			)
			if err != nil {
				logger.I("failed to create client", log.Error(err))
			} else {
				logger.D("client created")
				ag.SetClient(cl)

				err = func() error {
					dialCtx, cancelDial := context.WithTimeout(appCtx, config.Connectivity.DialTimeout)
					defer cancelDial()

					logger.I("establishing connectivity")
					return cl.Connect(dialCtx)
				}()

				if err != nil {
					logger.I("failed to establish connectivity", log.Error(err))
				} else {
					// start to sync
					logger.I("connected, starting communication")
					err = cl.Start(appCtx)
					if err != nil {
						logger.I("failed to communicate", log.Error(err))
					}
				}

				_ = cl.Close()
				logger.I("connectivity lost")
			}

			if err != nil {
				wait := bs.Next(id)

				logger.I("connectivity backoff", log.Duration("wait", wait))

				backoffTimer.Reset(wait)
				select {
				case <-appCtx.Done():
					return nil
				case <-backoffTimer.C:
				}
			} else {
				bs.Reset(id)
			}
		}
	}

	return nil
}
