// +build !noconfhelper_pprof

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

package agent

import (
	"context"
	"errors"
	"net"
	"net/http"

	"arhat.dev/pkg/perfhelper"
)

type agentComponentPProf struct{}

func (c *agentComponentPProf) init(ctx context.Context, config *perfhelper.PProfConfig) {
	h := config.CreateHTTPHandlerIfEnabled(false)

	mux := http.NewServeMux()
	basePath := config.HTTPPath

	mux.Handle(basePath, http.StripPrefix(basePath, h))

	srv := &http.Server{
		Handler: mux,
		Addr:    config.Listen,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()
}
