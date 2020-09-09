// +build !nometrics

/*
The MIT License (MIT)

Copyright (c) 2016 Martin Lindhe

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package windowsexporter

import (
	"fmt"
	"sync"
	"time"

	"github.com/prometheus-community/windows_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	//"github.com/prometheus/common/log"
)

// WmiCollector implements the prometheus.Collector interface.
type (
	collectorOutcome int

	WmiCollector struct {
		maxScrapeDuration time.Duration
		collectors        map[string]collector.Collector
	}
)

const (
	pending collectorOutcome = iota
	success
	failed
)

var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, "exporter", "collector_duration_seconds"),
		"wmi_exporter: Duration of a collection.",
		[]string{"collector"},
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, "exporter", "collector_success"),
		"wmi_exporter: Whether the collector was successful.",
		[]string{"collector"},
		nil,
	)
	scrapeTimeoutDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, "exporter", "collector_timeout"),
		"wmi_exporter: Whether the collector timed out.",
		[]string{"collector"},
		nil,
	)
	snapshotDuration = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, "exporter", "perflib_snapshot_duration_seconds"),
		"Duration of perflib snapshot capture",
		nil,
		nil,
	)

	// This can be removed when client_golang exposes this on Windows
	// (See https://github.com/prometheus/client_golang/issues/376)
	startTime     = float64(time.Now().Unix())
	startTimeDesc = prometheus.NewDesc(
		"process_start_time_seconds",
		"Start time of the process since unix epoch in seconds.",
		nil,
		nil,
	)
)

// Describe sends all the descriptors of the collectors included to
// the provided channel.
func (coll WmiCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc
}

// Collect sends the collected metrics from each of the collectors to
// prometheus.
func (coll WmiCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		startTimeDesc,
		prometheus.CounterValue,
		startTime,
	)

	t := time.Now()
	cs := make([]string, 0, len(coll.collectors))
	for name := range coll.collectors {
		cs = append(cs, name)
	}
	scrapeContext, err := collector.PrepareScrapeContext(cs)
	ch <- prometheus.MustNewConstMetric(
		snapshotDuration,
		prometheus.GaugeValue,
		time.Since(t).Seconds(),
	)
	if err != nil {
		ch <- prometheus.NewInvalidMetric(scrapeSuccessDesc, fmt.Errorf("failed to prepare scrape: %v", err))
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(len(coll.collectors))
	collectorOutcomes := make(map[string]collectorOutcome)
	for name := range coll.collectors {
		collectorOutcomes[name] = pending
	}

	metricsBuffer := make(chan prometheus.Metric)
	l := sync.Mutex{}
	finished := false
	go func() {
		for m := range metricsBuffer {
			l.Lock()
			if !finished {
				ch <- m
			}
			l.Unlock()
		}
	}()

	for name, c := range coll.collectors {
		go func(name string, c collector.Collector) {
			defer wg.Done()
			outcome := execute(name, c, scrapeContext, metricsBuffer)
			l.Lock()
			if !finished {
				collectorOutcomes[name] = outcome
			}
			l.Unlock()
		}(name, c)
	}

	allDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(allDone)
		close(metricsBuffer)
	}()

	// Wait until either all collectors finish, or timeout expires
	select {
	case <-allDone:
	case <-time.After(coll.maxScrapeDuration):
	}

	l.Lock()
	finished = true

	remainingCollectorNames := make([]string, 0)
	for name, outcome := range collectorOutcomes {
		var successValue, timeoutValue float64
		if outcome == pending {
			timeoutValue = 1.0
			remainingCollectorNames = append(remainingCollectorNames, name)
		}
		if outcome == success {
			successValue = 1.0
		}

		ch <- prometheus.MustNewConstMetric(
			scrapeSuccessDesc,
			prometheus.GaugeValue,
			successValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			scrapeTimeoutDesc,
			prometheus.GaugeValue,
			timeoutValue,
			name,
		)
	}

	if len(remainingCollectorNames) > 0 {
		//log.Warn("Collection timed out, still waiting for ", remainingCollectorNames)
	}

	l.Unlock()
}

func execute(name string, c collector.Collector, ctx *collector.ScrapeContext, ch chan<- prometheus.Metric) collectorOutcome {
	t := time.Now()
	err := c.Collect(ctx, ch)
	duration := time.Since(t).Seconds()
	ch <- prometheus.MustNewConstMetric(
		scrapeDurationDesc,
		prometheus.GaugeValue,
		duration,
		name,
	)

	if err != nil {
		//log.Errorf("collector %s failed after %fs: %s", name, duration, err)
		return failed
	}
	//log.Debugf("collector %s succeeded after %fs.", name, duration)
	return success
}
