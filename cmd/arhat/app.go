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
	"context"
	"fmt"
	"sort"
	"time"

	"arhat.dev/pkg/backoff"
	"arhat.dev/pkg/log"
	"github.com/spf13/pflag"

	"arhat.dev/arhat/pkg/agent"
	"arhat.dev/arhat/pkg/client"
	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/arhat/pkg/constant"
)

var (
	appCtx       context.Context
	configFile   string
	config       = new(conf.Config)
	cliLogConfig = new(log.Config)

	flags = pflag.NewFlagSet("arhat", pflag.ContinueOnError)
)

func init() {
	flags.StringVarP(
		&configFile,
		"config",
		"c",
		constant.DefaultArhatConfigFile,
		"path to the arhat config file",
	)

	flags.StringVar(
		&cliLogConfig.Level,
		"log.level",
		"error",
		"log level, one of [verbose, debug, info, error, silent]",
	)

	flags.StringVar(
		&cliLogConfig.Format,
		"log.format",
		"console",
		"log output format, one of [console, json]",
	)

	flags.StringVar(
		&cliLogConfig.File,
		"log.file",
		"stderr",
		"log to this file",
	)

	flags.AddFlagSet(conf.FlagsForArhatHostConfig("host.", &config.Arhat.Host))

	flags.AddFlagSet(conf.FlagsForExtensionConfig("ext.", &config.Extension))
}

func runApp(appCtx context.Context, config *conf.Config) error {
	logger := log.Log.WithName("cmd")

	ag, err := agent.NewAgent(appCtx, logger.WithName("agent"), config)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	bs := backoff.NewStrategy(
		config.Connectivity.InitialBackoff,
		config.Connectivity.MaxBackoff,
		config.Connectivity.BackoffFactor,
		1,
	)

	// create a stopped timer
	backoffTimer := time.NewTimer(0)
	if !backoffTimer.Stop() {
		<-backoffTimer.C
	}
	defer backoffTimer.Stop()

	methods := config.Connectivity.Methods
	sort.SliceStable(methods, func(i, j int) bool {
		return methods[i].Priority < methods[j].Priority
	})

	// chroot into new rootfs after extension hub listened on the original host
	if config.Arhat.Chroot != "" {
		err = chroot(config.Arhat.Chroot)
		if err != nil {
			return fmt.Errorf("failed to chroot to %q: %w", config.Arhat.Chroot, err)
		}
	}

	for {
		// select connectivity according to its priority
		for i := 0; i < len(config.Connectivity.Methods); i++ {
			select {
			case <-appCtx.Done():
				// application exited, no more retry
				return nil
			default:
				// not exited, maintain connectivity
			}

			name := config.Connectivity.Methods[i].Name
			id := fmt.Sprintf("%s@%d", name, i)

			logger.I("creating client", log.String("id", id))
			cl, err := client.NewClient(
				appCtx, name, ag.HandleCmd, config.Connectivity.Methods[i].Config,
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

	// unreachable
}
