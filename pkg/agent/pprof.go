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
	"strings"

	"arhat.dev/pkg/perfhelper"
)

type agentComponentPProf struct{}

func (c *agentComponentPProf) init(ctx context.Context, config *perfhelper.PProfConfig) error {
	handlers := config.CreateHTTPHandlersIfEnabled()
	if handlers == nil {
		// no enabled
		return nil
	}

	tlsConfig, err := config.TLS.GetTLSConfig(true)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()

	basePath := strings.TrimRight(config.HTTPPath, "/")
	for suffix, h := range handlers {
		mux.Handle(basePath+"/"+suffix, h)
	}

	srv := &http.Server{
		Handler:   mux,
		Addr:      config.Listen,
		TLSConfig: tlsConfig,
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

	return nil
}
