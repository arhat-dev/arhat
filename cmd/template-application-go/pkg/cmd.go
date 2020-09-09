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

package pkg

import (
	"context"

	"arhat.dev/pkg/log"
	"github.com/spf13/cobra"

	"arhat.dev/template-application-go/pkg/conf"
	"arhat.dev/template-application-go/pkg/constant"
)

func NewTemplateApplicationGoCmd() *cobra.Command {
	var (
		appCtx       context.Context
		configFile   string
		config       = new(conf.TemplateApplicationGoConfig)
		cliLogConfig = new(log.Config)
	)

	templateApplicationGoCmd := &cobra.Command{
		Use:           "template-application-go",
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

	flags := templateApplicationGoCmd.PersistentFlags()

	flags.StringVarP(&configFile, "config", "c", constant.DefaultTemplateApplicationGoConfigFile,
		"path to the templateApplicationGo config file")
	flags.AddFlagSet(conf.FlagsForTemplateApplicationGo("", &config.TemplateApplicationGo))

	return templateApplicationGoCmd
}

func run(appCtx context.Context, config *conf.TemplateApplicationGoConfig) error {
	logger := log.Log.WithName("TemplateApplicationGo")

	_ = logger
	return nil
}
