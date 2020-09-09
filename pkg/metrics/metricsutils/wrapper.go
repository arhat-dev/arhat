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

package metricsutils

import (
	"bytes"
	"compress/gzip"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

func CreateMetricsCollectingFunc(g prometheus.Gatherer) func() ([]byte, error) {
	var (
		gzipWriter = gzip.NewWriter(nil)
		gzipMu     = new(sync.Mutex)
	)

	return func() ([]byte, error) {
		gzipMu.Lock()
		defer gzipMu.Unlock()

		mfs, err := g.Gather()
		if err != nil {
			return nil, err
		}

		buf := &bytes.Buffer{}
		gzipWriter.Reset(buf)
		defer func() { _ = gzipWriter.Close() }()

		encoder := expfmt.NewEncoder(gzipWriter, expfmt.FmtProtoDelim)
		for _, mf := range mfs {
			if err := encoder.Encode(mf); err != nil {
				return nil, err
			}
		}

		_ = gzipWriter.Close()
		return buf.Bytes(), nil
	}
}
