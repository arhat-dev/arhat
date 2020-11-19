// +build !nometrics
// +build !windows

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

package unixexporter

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"arhat.dev/aranya-proto/aranyagopb"
	"github.com/prometheus/client_golang/prometheus"
	promlog "github.com/prometheus/common/log"
	"github.com/prometheus/node_exporter/collector"
	"gopkg.in/alecthomas/kingpin.v2"

	"arhat.dev/arhat/pkg/metrics/metricsutils"
)

func init() {
	// workaround for node_exporter
	kingpin.CommandLine.ErrorWriter(ioutil.Discard)
	kingpin.CommandLine.UsageWriter(ioutil.Discard)
	kingpin.CommandLine.Terminate(nil)

	_ = promlog.Base().SetLevel("panic")
}

type logWrapper struct {
}

func (l *logWrapper) Log(kv ...interface{}) error {
	return nil
}

func CreateGatherer(config *aranyagopb.MetricsConfigCmd) (prometheus.Gatherer, error) {
	args := []string{os.Args[0]}

	var collectors []string
	enableGoCollector, enabledCollectors := metricsutils.GetEnabledCollectors(config.Collect)
	for c := range enabledCollectors {
		collectors = append(collectors, c)
		args = append(args, fmt.Sprintf("--collector.%s", c))
	}

	extraArgs, err := metricsutils.GetCollectorExtraArgs(enabledCollectors, config.ExtraArgs)
	if err != nil {
		return nil, err
	}

	args = append(args, extraArgs...)

	osArgs := os.Args
	os.Args = args
	_ = kingpin.Parse()
	os.Args = osArgs

	logger := &logWrapper{}
	nc, err := collector.NewNodeCollector(logger, collectors...)
	if err != nil {
		return nil, err
	}

	registry := prometheus.NewRegistry()
	err = registry.Register(
		prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: "node_exporter",
				Name:      "build_info",
				Help: fmt.Sprintf(
					// nolint:lll
					"A metric with a constant '1' value labeled by version, revision, branch, and goversion from which %s was built.",
					"node_exporter",
				),
				ConstLabels: prometheus.Labels{
					"version":   "v1.0.1",
					"revision":  "3715be6ae899f2a9b9dbfd9c39f3e09a7bd4559f",
					"branch":    "HEAD",
					"goversion": runtime.Version(),
				},
			},
			func() float64 { return 1 },
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register version collector: %w", err)
	}

	err = registry.Register(
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register process collector: %w", err)
	}

	if enableGoCollector {
		err = registry.Register(prometheus.NewGoCollector())
		if err != nil {
			return nil, fmt.Errorf("failed to register go collector: %w", err)
		}
	}

	if err := registry.Register(nc); err != nil {
		return nil, fmt.Errorf("failed to register node exporter collector: %w", err)
	}

	return registry, nil
}
