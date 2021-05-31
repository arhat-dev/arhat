// +build !nometrics
// +build windows

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

package windowsexporter

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"time"

	"arhat.dev/aranya-proto/aranyagopb"
	"github.com/prometheus-community/windows_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	promlog "github.com/prometheus/common/log"
	"gopkg.in/alecthomas/kingpin.v2"

	"arhat.dev/arhat/pkg/metrics/metricsutils"
)

func init() {
	// workaround for windows_exporter
	kingpin.CommandLine.ErrorWriter(ioutil.Discard)
	kingpin.CommandLine.UsageWriter(ioutil.Discard)
	kingpin.CommandLine.Terminate(nil)

	_ = promlog.Base().SetLevel("panic")
}

func CreateGatherer(config *aranyagopb.MetricsConfigCmd) (prometheus.Gatherer, error) {
	var (
		collectors = make(map[string]collector.Collector)
		args       = []string{os.Args[0]}
	)

	enableGoCollector, enabledCollectors := metricsutils.GetEnabledCollectors(config.Collect)

	for name := range enabledCollectors {
		c, err := collector.Build(name)
		if err != nil {
			return nil, fmt.Errorf("failed to build collector '%s': %w", name, err)
		}

		collectors[name] = c
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

	registry := prometheus.NewRegistry()
	err = registry.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: "windows_exporter",
			Name:      "build_info",
			Help: fmt.Sprintf(
				// nolint:lll
				"A metric with a constant '1' value labeled by version, revision, branch, and goversion from which %s was built.",
				"windows_exporter",
			),
			ConstLabels: prometheus.Labels{
				"version":   "v0.16.0",
				"revision":  "f316d81d50738eb0410b0748c5dcdc6874afe95a",
				"branch":    "HEAD",
				"goversion": runtime.Version(),
			},
		},
		func() float64 { return 1 },
	))
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

	windowsColl := &windowsCollector{
		collectors:        collectors,
		maxScrapeDuration: 20 * time.Second,
	}

	if err := registry.Register(windowsColl); err != nil {
		return nil, err
	}

	return registry, nil
}
