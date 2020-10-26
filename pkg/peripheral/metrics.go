// +build !noperipheral
// +build !noperipheral_metrics

package peripheral

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
	"time"

	dto "github.com/prometheus/client_model/go"

	"arhat.dev/arhat/pkg/metrics/metricsutils"
)

func NewMetricsCache(maxCacheTime time.Duration) *MetricsCache {
	return &MetricsCache{
		maxCacheTime: int64(maxCacheTime),
		metricsCache: make(map[string]*dto.MetricFamily),
		cacheMU:      new(sync.Mutex),
	}
}

type MetricsCache struct {
	maxCacheTime int64
	metricsCache map[string]*dto.MetricFamily
	cacheMU      *sync.Mutex
}

func (m *MetricsCache) CacheMetrics(metrics []*dto.MetricFamily) {
	m.cacheMU.Lock()
	defer m.cacheMU.Unlock()

	lastValidTime := (time.Now().UnixNano() - m.maxCacheTime) / 1000000
	for _, mf := range metrics {
		if mf == nil || mf.Name == nil {
			continue
		}

		f, ok := m.metricsCache[*mf.Name]
		if !ok {
			f = mf
		}

		f.Metric = append(f.Metric, mf.Metric...)
		sort.Slice(f.Metric, func(i, j int) bool {
			return *f.Metric[i].TimestampMs < *f.Metric[j].TimestampMs
		})

		for i := range f.Metric {
			if *f.Metric[i].TimestampMs > lastValidTime {
				f.Metric = f.Metric[i:]
			}
		}

		m.metricsCache[*mf.Name] = f
	}
}

func (m *MetricsCache) RetrieveCachedMetrics() []*dto.MetricFamily {
	m.cacheMU.Lock()
	defer m.cacheMU.Unlock()

	var result []*dto.MetricFamily
	for n := range m.metricsCache {
		result = append(result, m.metricsCache[n])
	}

	m.metricsCache = make(map[string]*dto.MetricFamily)

	return result
}

func (m *Manager) CacheMetrics(metrics []*dto.MetricFamily) {
	m.metricsCache.CacheMetrics(metrics)
}

func (m *Manager) RetrieveCachedMetrics() []*dto.MetricFamily {
	return m.metricsCache.RetrieveCachedMetrics()
}

// nolint:gocyclo
func (m *Manager) CollectMetrics(peripheralIDs ...string) (
	metricsForNode []*dto.MetricFamily,
	paramsForAgent []map[string]string,
	metricsForAgent [][]*dto.MetricFamily,
) {
	var (
		wg                                = new(sync.WaitGroup)
		reportViaStandaloneClientResultCh = make(chan map[MetricReportKey]map[string]*MetricReportSpec)
		reportViaNodeMetricsResultCh      = make(chan map[MetricReportKey]map[string]*MetricReportSpec)
		reportViaAgentClientResultCh      = make(chan map[MetricReportKey]map[string]*MetricReportSpec)
	)

	m.mu.RLock()
	var peripherals []*Peripheral
	all := len(peripheralIDs) == 0
	for _, id := range peripheralIDs {
		dev, ok := m.peripherals[id]
		if !ok && !all {
			continue
		}

		peripherals = append(peripherals, dev)
	}
	m.mu.RUnlock()

	peripheralCount := len(peripherals)
	for _, dev := range peripherals {
		resultCh := make(chan *Metric)

		// map this metrics collecting tasks to workers

		wg.Add(1)
		go func(dev *Peripheral) {
			defer wg.Done()

			go dev.CollectMetrics(resultCh)

			// key: report key
			// value key: metric name
			// value value: metric report spec
			var (
				reportViaStandaloneClient = make(map[MetricReportKey]map[string]*MetricReportSpec)
				reportViaNodeMetrics      = make(map[MetricReportKey]map[string]*MetricReportSpec)
				reportViaAgentClient      = make(map[MetricReportKey]map[string]*MetricReportSpec)
			)

			for result := range resultCh {

				target := reportViaNodeMetrics
				switch {
				case result.ReportKey.ReporterName != "" && result.ReportKey.ParamsHashHex != "":
					target = reportViaStandaloneClient
				case result.ReportKey.ReporterName == "" && result.ReportKey.ParamsHashHex != "":
					target = reportViaAgentClient
				}

				mc, ok := target[result.ReportKey]
				if !ok {
					mc = make(map[string]*MetricReportSpec)
				}

				report, ok := mc[result.Name]
				if !ok {
					report = &MetricReportSpec{
						Metrics: &dto.MetricFamily{
							Name: &result.Name,
							Type: result.ValueType.Enum(),
						},

						ParamsForReporting: result.ParamsForReporting,
					}
				}

				mtc := &dto.Metric{
					TimestampMs: &result.Timestamp,
				}

				switch result.ValueType {
				case dto.MetricType_COUNTER:
					mtc.Counter = &dto.Counter{
						Value: &result.Value,
					}
				case dto.MetricType_GAUGE:
					mtc.Gauge = &dto.Gauge{
						Value: &result.Value,
					}
				case dto.MetricType_UNTYPED:
					mtc.Untyped = &dto.Untyped{
						Value: &result.Value,
					}
				}

				report.Metrics.Metric = append(report.Metrics.Metric, mtc)

				mc[result.Name] = report
				target[result.ReportKey] = mc
			}

			select {
			case reportViaStandaloneClientResultCh <- reportViaStandaloneClient:
			case <-m.ctx.Done():
				return
			}

			select {
			case reportViaNodeMetricsResultCh <- reportViaNodeMetrics:
			case <-m.ctx.Done():
				return
			}

			select {
			case reportViaAgentClientResultCh <- reportViaAgentClient:
			case <-m.ctx.Done():
				return
			}
		}(dev)
	}

	var (
		reportViaStandaloneClient = make(map[MetricReportKey]map[string]*MetricReportSpec)
		reportViaNodeMetrics      = make(map[MetricReportKey]map[string]*MetricReportSpec)
		reportViaAgentClient      = make(map[MetricReportKey]map[string]*MetricReportSpec)
	)

	for i := 0; i < peripheralCount; i++ {
		select {
		case m := <-reportViaStandaloneClientResultCh:
			reportViaStandaloneClient = mergeCollectedMetrics(reportViaStandaloneClient, m)
		case m := <-reportViaNodeMetricsResultCh:
			reportViaNodeMetrics = mergeCollectedMetrics(reportViaNodeMetrics, m)
		case m := <-reportViaAgentClientResultCh:
			reportViaAgentClient = mergeCollectedMetrics(reportViaAgentClient, m)
		case <-m.ctx.Done():
			return
		}
	}

	wg.Wait()

	close(reportViaStandaloneClientResultCh)
	close(reportViaNodeMetricsResultCh)
	close(reportViaAgentClientResultCh)

	reporterHashes, params, metrics := normalizeCollectedMetrics(reportViaStandaloneClient)
	reporters := make([]*MetricsReporter, 0, len(reporterHashes))

	m.mu.RLock()
	for i, h := range reporterHashes {
		r, ok := m.metricsReporters[h]
		if !ok {
			params = append(params[:i], params[i+1:]...)
			metrics = append(metrics[:i], metrics[i+1:]...)

			// TODO: log error
			continue
		}

		reporters = append(reporters, r)
	}
	m.mu.Unlock()

	for i, r := range reporters {
		err := r.ReportMetrics(params[i], metrics[i])
		if err != nil {
			// TODO: log error
			_ = err
		}
	}

	_, paramsForAgent, metricsForAgent = normalizeCollectedMetrics(reportViaAgentClient)

	var mn [][]*dto.MetricFamily
	_, _, mn = normalizeCollectedMetrics(reportViaNodeMetrics)
	for i := range mn {
		metricsForNode = append(metricsForNode, mn[i]...)
	}

	return
}

type Metric struct {
	Name      string
	Value     float64
	Timestamp int64
	ValueType dto.MetricType

	ReportKey          MetricReportKey
	ParamsForReporting map[string]string
}

// MetricReportSpec defines how to report this metric
type MetricReportSpec struct {
	Metrics *dto.MetricFamily

	ParamsForReporting map[string]string
}

func (r *MetricsReporter) ReportMetrics(params map[string]string, metrics []*dto.MetricFamily) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	buf := new(bytes.Buffer)
	err := metricsutils.EncodeMetrics(buf, metrics)
	if err != nil {
		return fmt.Errorf("failed to encode metrics: %w", err)
	}

	_, err = r.conn.Operate(r.ctx, params, buf.Bytes())
	return err
}

func mergeCollectedMetrics(
	existing, newValues map[MetricReportKey]map[string]*MetricReportSpec,
) map[MetricReportKey]map[string]*MetricReportSpec {
	for k, mc := range newValues {
		existingMC, ok := existing[k]
		if !ok {
			existing[k] = newValues[k]
			continue
		}

		for name, ms := range mc {
			existingMS, ok := existingMC[name]
			if !ok {
				existingMC[name] = mc[name]
				continue
			}

			existingMS.Metrics.Metric = append(existingMS.Metrics.Metric, ms.Metrics.Metric...)

			existingMC[name] = existingMS
		}

		existing[k] = existingMC
	}

	return existing
}

func normalizeCollectedMetrics(
	mrs map[MetricReportKey]map[string]*MetricReportSpec,
) (reporterHashHexes []string, params []map[string]string, metrics [][]*dto.MetricFamily) {
	for k, ms := range mrs {
		// share same reporter and params
		var (
			p          map[string]string
			collection []*dto.MetricFamily
		)

		for _, mr := range ms {
			if len(mr.Metrics.Metric) == 0 {
				continue
			}

			sort.Slice(mr.Metrics.Metric, func(i, j int) bool {
				// nolint:scopelint
				return *mr.Metrics.Metric[i].TimestampMs < *mr.Metrics.Metric[j].TimestampMs
			})

			collection = append(collection, mr.Metrics)

			p = mr.ParamsForReporting
		}

		reporterHashHexes = append(reporterHashHexes, k.ReporterName)
		params = append(params, p)
		metrics = append(metrics, collection)
	}

	return
}
