package main

import (
	"context"
	"fmt"
	"runtime"
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
	runtime.GOMAXPROCS(config.Arhat.Optimization.MaxProcessors)

	logger := log.Log.WithName("cmd")

	// // handle pprof
	// if cfg := config.Arhat.Optimization.PProf; cfg.Enabled {
	// 	mgr, err := manager.NewPProfManager(cfg.Listen, cfg.HTTPPath, cfg.MutexProfileFraction, cfg.BlockProfileRate)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to listen tcp for pprof server: %w", err)
	// 	}

	// 	logger.D("serving pprof stats")
	// 	go func() {
	// 		if err := mgr.Start(); err != nil && err != http.ErrServerClosed {
	// 			logger.E("failed to start pprof server", log.Error(err))
	// 			os.Exit(1)
	// 		}
	// 	}()
	// }

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
			cl, err := client.NewConnectivityClient(
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
