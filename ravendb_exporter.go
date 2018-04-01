package main

import (
	"time"
	//"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	namespace = "ravendb"
	subsystem = "exporter"
)

var (
	ravenBaseURL = "http://localhost:8080"
	timeout      = time.Second * 10
)

type exporter struct {
	up         prometheus.Gauge
	workingSet prometheus.Gauge
}

func newExporter() *exporter {
	return &exporter{
		up:         createGauge("up", "Whether the RavenDB scrape was successful"),
		workingSet: createGauge("ravendb_working_set_bytes", "Process working set"),
	}
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up.Desc()
	ch <- e.workingSet.Desc()
}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	if stats, err := getStats(); err != nil {
		e.up.Set(0)
		ch <- e.up
	} else {
		e.up.Set(1)
		ch <- e.up

		e.workingSet.Set(stats.memory["WorkingSet"].(float64))
		ch <- e.workingSet
	}
}

func createGauge(name string, help string) prometheus.Gauge {
	return prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
	})
}

func serveLandingPage() {
	var landingPage = []byte(`<html>
		<head><title>RavenDB exporter</title></head>
		<body>
		<h1>RavenDB exporter</h1>
		<p><a href='/metrics'>Metrics</a></p>
		</body>
		</html>
		`)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(landingPage) // nolint: errcheck
	})
}

func serveMetrics() {
	exporter := newExporter()
	prometheus.MustRegister(exporter)

	http.Handle("/metrics", promhttp.Handler())
}

func main() {

	serveLandingPage()
	serveMetrics()

	log.Fatal(http.ListenAndServe(":9999", nil))
}
