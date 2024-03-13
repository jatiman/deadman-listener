package deadman

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
)

var (
	ticksTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "deadman_ticks_total",
			Help: "The total ticks passed in this listener",
		},
	)

	ticksNotified = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "deadman_ticks_notified",
			Help: "The number of ticks where notifications were sent.",
		},
	)

	failedNotifications = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "deadman_notifications_failed",
			Help: "The number of failed notifications.",
		},
	)
)

func init() {
	prometheus.MustRegister(
		ticksTotal,
		ticksNotified,
		failedNotifications,
	)
}

type Deadman struct {
	pinger           <-chan time.Time
	interval         time.Duration
	ticker           *time.Ticker
	closer           chan struct{}
	notifier         func() error
	logger           log.Logger
	additionalLabels model.LabelSet
}

func NewDeadMan(pinger <-chan time.Time, interval time.Duration, amURL string, logger log.Logger, additionalLabels model.LabelSet) (*Deadman, error) {
	return newDeadMan(pinger, interval, amNotifier(amURL, additionalLabels), logger, additionalLabels), nil
}

func newDeadMan(pinger <-chan time.Time, interval time.Duration, notifier func() error, logger log.Logger, additionalLabels model.LabelSet) *Deadman {
	return &Deadman{
		pinger:           pinger,
		interval:         interval,
		notifier:         notifier,
		closer:           make(chan struct{}),
		logger:           logger,
		additionalLabels: additionalLabels,
	}
}

func (d *Deadman) Run() error {
	d.ticker = time.NewTicker(d.interval)

	skip := true

	for {
		select {
		case <-d.ticker.C:
			ticksTotal.Inc()

			if !skip {
				ticksNotified.Inc()
				level.Warn(d.logger).Log("msg", "no heartbeat received within the time interval", "interval", d.interval)
				if d.notifier != nil {
					if err := d.notifier(); err != nil {
						failedNotifications.Inc()
						logError(d.logger, err, "operation", "notifier")
					}
				}
			}
			skip = false

		case <-d.pinger:
			skip = true

		case <-d.closer:
			if d.ticker != nil {
				d.ticker.Stop()
			}
			close(d.closer)
			return nil
		}
	}
}

func (d *Deadman) Stop() {
	if d.ticker != nil {
		d.ticker.Stop()
	}

	d.closer <- struct{}{}
}

func amNotifier(amURL string, additionalLabels model.LabelSet) func() error {
	alerts := []*model.Alert{{
		Labels: model.LabelSet{
			model.LabelName("alertname"): model.LabelValue("PrometheusAlertPipelineFailed"),
		}.Merge(additionalLabels),
		Annotations: model.LabelSet{
			model.LabelName("description"): model.LabelValue("Alertmanager does not receive any Watchdog/Deadman alert heartbeat. Please check the connectivity between Prometheus and Alertmanager"),
		},
	}}

	b, err := json.Marshal(alerts)

	if err != nil {
		return func() error {
			return fmt.Errorf("error in json.Marshal: %v", err)
		}
	}

	return func() error {
		client := &http.Client{}
		resp, err := client.Post(amURL, "application/json", bytes.NewReader(b))
		if err != nil {
			return fmt.Errorf("error in HTTP post: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode/100 != 2 {
			return fmt.Errorf("bad response status %v", resp.Status)
		}

		return nil
	}
}

func logError(logger log.Logger, err error, keyvals ...interface{}) {
	level.Error(logger).Log(append(keyvals, "err", err.Error())...)
}
