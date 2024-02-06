package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/jatiman/deadman-listener/deadman"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"
	"github.com/prometheus/common/promlog"
	promlogflag "github.com/prometheus/common/promlog/flag"
)

var (
	amURL            string
	interval         model.Duration
	logLevel         promlog.AllowedLevel
	httpPort         string
	additionalLabels string
)

func main() {

	app := kingpin.New(filepath.Base(os.Args[0]), "A deadman listener for Prometheus Alertmanager compatible notifications.")
	app.HelpFlag.Short('h')

	app.Flag("am.url", "The URL to POST alerts to.").
		Default("http://localhost:9093/api/v1/alerts").StringVar(&amURL)
	app.Flag("deadman.interval", "The heartbeat interval. An alert is sent if no heartbeat is sent.").
		Default("30s").SetValue(&interval)
	app.Flag("deadman.port", "The HTTP port that will be used by Deadman.").
		Default("9095").StringVar(&httpPort)
	app.Flag("alert.labels", "Additional labels to be added to alerts.").
		PlaceHolder("key1=value1,key2=value2,...").
		StringVar(&additionalLabels)

	promlogConfig := &promlog.Config{Level: &logLevel}
	promlogflag.AddFlags(app, promlogConfig)

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Error parsing commandline arguments"))
		app.Usage(os.Args[1:])
		os.Exit(2)
	}

	banner()
	logger := promlog.New(promlogConfig)

	pinger := make(chan time.Time)
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/ping", simpleHandler(pinger, logger))
	go http.ListenAndServe(":"+httpPort, nil)

	d, err := deadman.NewDeadMan(pinger, time.Duration(interval), amURL, log.With(logger, "component", "deadman"), parseAdditionalLabels(additionalLabels))
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(2)
	}

	d.Run()
}

func simpleHandler(pinger chan<- time.Time, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Logger.Log(logger, "method", r.Method, "path", r.URL.Path)

		pinger <- time.Now()
		fmt.Fprint(w, "pong")
	}
}

func parseAdditionalLabels(labelsStr string) model.LabelSet {
	r := make(model.LabelSet)

	keyvals := strings.Split(labelsStr, ",")
	for _, keyval := range keyvals {
		kv := strings.SplitN(keyval, "=", 2)
		if len(kv) == 2 {
			r[model.LabelName(kv[0])] = model.LabelValue(kv[1])
		}
	}

	return r
}

func banner() {
	fmt.Println("--------------------------------")
	fmt.Println("Deadman Listener is starting ...")
	fmt.Printf("Deadman Listener alerts to %s if no heartbeat for more than %s\n", amURL, interval)
	fmt.Printf("Deadman Listener started on port %s\n", httpPort)
	fmt.Println("Access Deadman Listener's prometheus metrics at /metrics and the heartbeat at /ping")
	fmt.Println("--------------------------------")
}
