package main

import (
	"github.com/prometheus/client_golang/prometheus"
	jp "github.com/buger/jsonparser"
)

const (
	namespace = "ravendb"
	subsystem = "exporter"
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
	log.Info("Running scrape")

	if stats, err := getStats(); err != nil {
		log.WithError(err).Error("Error while getting data from RavenDB")

		e.up.Set(0)
		ch <- e.up
	} else {
		e.up.Set(1)
		ch <- e.up

		value, _ := jp.GetFloat(stats.memory, "WorkingSet")
		e.workingSet.Set(value)
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
