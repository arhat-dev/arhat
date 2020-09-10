// +build !nometrics

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
	"os"
	"time"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/arhat/pkg/metrics/metricsutils"
	"arhat.dev/arhat/pkg/types"
	"github.com/prometheus-community/windows_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

func CreateNodeMetricsGatherer(config *aranyagopb.MetricsConfigOptions) (types.MetricsCollectFunc, error) {
	var (
		collectors        = make(map[string]collector.Collector)
		enabledCollectors = metricsutils.GetEnabledCollectors(config.Collect)
		args              = []string{os.Args[0]}
	)

	for name := range enabledCollectors {
		c, err := collector.Build(name)
		if err != nil {
			return nil, fmt.Errorf("failed to build collector '%s': %w", name, err)
		}

		collectors[name] = c
	}

	extraArgs, err := metricsutils.GetExtraArgs(enabledCollectors, config.ExtraArgs)
	if err != nil {
		return nil, err
	}

	args = append(args, extraArgs...)

	osArgs := os.Args
	os.Args = args
	_ = kingpin.Parse()
	os.Args = osArgs

	registry := prometheus.NewRegistry()
	registry.MustRegister(
		version.NewCollector("wmi_exporter"),
		prometheus.NewGoCollector(),
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
	)

	wmiCollector := &WmiCollector{
		collectors:        collectors,
		maxScrapeDuration: 20 * time.Second,
	}

	if err := registry.Register(wmiCollector); err != nil {
		return nil, err
	}

	return metricsutils.CreateMetricsCollectingFunc(prometheus.Gatherers{registry}), nil
}
