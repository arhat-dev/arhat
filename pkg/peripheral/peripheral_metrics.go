// +build !noperipheral
// +build !noperipheral_metrics

package peripheral

import (
	"sync"
	"time"

	"arhat.dev/aranya-proto/aranyagopb"
	dto "github.com/prometheus/client_model/go"
)

// CollectMetrics will collect all metrics configured when creating this peripheral and close the resultCh
// when finished
func (d *Peripheral) CollectMetrics(resultCh chan<- *Metric) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.metrics) == 0 {
		return
	}

	workers := 1
	if len(d.metrics) > 1 {
		workers = len(d.metrics) / 5
	}

	if workers > 5 {
		workers = 5
	}

	var (
		wg        = new(sync.WaitGroup)
		collectCh = make(chan *MetricSpec, 1)
	)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for spec := range collectCh {
				metricValues, err := d.conn.CollectMetrics(d.ctx, spec.ParamsForCollecting)
				if err != nil {
					// TODO: log error
					continue
				}

				for _, mv := range metricValues {
					value, err := convertMetricValue(mv.Value)
					if err != nil {
						// TODO: log error
						continue
					}

					ts := mv.Timestamp
					if ts == 0 {
						now := time.Now()
						ts = now.UnixNano()
					}

					valueType := dto.MetricType_UNTYPED
					switch spec.ValueType {
					case aranyagopb.METRICS_VALUE_TYPE_COUNTER:
						valueType = dto.MetricType_COUNTER
					case aranyagopb.METRICS_VALUE_TYPE_GAUGE:
						valueType = dto.MetricType_GAUGE
					}

					select {
					case resultCh <- &Metric{
						Name:      spec.Name,
						Value:     value,
						Timestamp: ts / 1000000,
						ValueType: valueType,

						ReportKey: spec.ReportKey,
					}:
					case <-d.ctx.Done():
						return
					}
				}
			}
		}()
	}

	for i := range d.metrics {
		select {
		case collectCh <- d.metrics[i]:
		case <-d.ctx.Done():
			return
		}
	}

	close(collectCh)
	wg.Wait()
	close(resultCh)
}
