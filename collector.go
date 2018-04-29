package main

import (
	"strconv"
	"regexp"
	"github.com/prometheus/client_golang/prometheus"
	jp "github.com/buger/jsonparser"
)

const (
	namespace = "ravendb"
	subsystem = ""
)

type exporter struct {
	up         					prometheus.Gauge
	workingSet 					prometheus.Gauge
	cpuTime						prometheus.Counter
	isLeader					prometheus.Gauge
	requestCount				prometheus.Counter
	documentPutCount			prometheus.Counter
	documentPutBytes			prometheus.Counter
	mapIndexIndexedCount		prometheus.Counter
	mapReduceIndexMappedCount	prometheus.Counter
	mapReduceIndexReducedCount	prometheus.Counter

	databaseDocumentCount		*prometheus.GaugeVec

}

func newExporter() *exporter {
	return &exporter{
		up:         					createGauge("up", "Whether the RavenDB scrape was successful"),
		workingSet: 					createGauge("working_set_bytes", "Process working set"),
		cpuTime:						createCounter("cpu_time_seconds", "CPU time"),
		isLeader:						createGauge("is_leader", "If 1, then node is the cluster leader, otherwise 0"),
		requestCount:					createCounter("request_count", "Server-wide request count"),
		documentPutCount:				createCounter("document_put_count", "Server-wide document puts count"),
		documentPutBytes:				createCounter("document_put_bytes", "Server-wide document put bytes"),
		mapIndexIndexedCount:			createCounter("mapindex_indexed_count", "Server-wide map index indexed count"),
		mapReduceIndexMappedCount:		createCounter("mapreduceindex_mapped_count", "Server-wide map-reduce index mapped count"),
		mapReduceIndexReducedCount:		createCounter("mapreduceindex_reduced_count", "Server-wide map-reduce index reduced count"),

		databaseDocumentCount:			createGaugeDatabaseVec("database_document_count", "Count of documents in a database"),
	}
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up.Desc()
	ch <- e.workingSet.Desc()
	ch <- e.cpuTime.Desc()
	ch <- e.isLeader.Desc()
	ch <- e.requestCount.Desc()
	ch <- e.documentPutCount.Desc()
	ch <- e.documentPutBytes.Desc()
	ch <- e.mapIndexIndexedCount.Desc()
	ch <- e.mapReduceIndexMappedCount.Desc()
	ch <- e.mapReduceIndexReducedCount.Desc()

	e.databaseDocumentCount.Describe(ch)
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

		e.workingSet.Set(getMemoryWorkingSet(stats))
		ch <- e.workingSet

		e.cpuTime.Set(getCPUTime(stats))
		ch <- e.cpuTime

		e.isLeader.Set(getIsLeader(stats))
		ch <- e.isLeader

		e.requestCount.Set(getRequestCount(stats))
		ch <- e.requestCount

		e.documentPutCount.Set(getDocumentPutCount(stats))
		ch <- e.documentPutCount

		e.documentPutBytes.Set(getDocumentPutBytesCount(stats))
		ch <- e.documentPutBytes

		e.mapIndexIndexedCount.Set(getMapIndexIndexedCount(stats))
		ch <- e.mapIndexIndexedCount

		e.mapReduceIndexMappedCount.Set(getMapReduceIndexMappedCount(stats))
		ch <- e.mapReduceIndexMappedCount

		e.mapReduceIndexReducedCount.Set(getMapReduceIndexReducedCount(stats))
		ch <- e.mapReduceIndexReducedCount

		for _, dbs := range stats.dbStats {
			e.databaseDocumentCount.WithLabelValues(dbs.database).Set(getDatabaseDocumentCount(dbs));
		}

		e.databaseDocumentCount.Collect(ch)
	}
}

func getCPUTime(stats *stats) float64 {
	var cpuTimeString string
	jp.ArrayEach(stats.cpu, func(value []byte, dataType jp.ValueType, offset int, err error) {
		cpuTimeString, _ = jp.GetString(value, "TotalProcessorTime") // just use the last entry in the array TODO: why is this an array?
	}, "CpuStats")

	return timeSpanToSeconds(cpuTimeString)
}

func getMemoryWorkingSet(stats *stats) float64 {
	value, _ := jp.GetFloat(stats.memory, "WorkingSet")
	return value
}

func getIsLeader(stats * stats) float64 {
	value, _ := jp.GetString(stats.nodeInfo, "CurrentState")
	if value == "Leader" {
		return 1
	} 
	return 0
}

func getRequestCount(stats *stats) float64 {
	value, _ := jp.GetFloat(stats.metrics, "Requests", "RequestsPerSec", "Count")
	return value
}

func getDocumentPutCount(stats *stats) float64 {
	value, _ := jp.GetFloat(stats.metrics, "Docs", "PutsPerSec", "Count")
	return value
}

func getDocumentPutBytesCount(stats *stats) float64 {
	value, _ := jp.GetFloat(stats.metrics, "Docs", "BytesPutsPerSec", "Count")
	return value
}

func getMapIndexIndexedCount(stats *stats) float64 {
	value, _ := jp.GetFloat(stats.metrics, "MapIndexes", "MappedPerSec", "Count")
	return value
}

func getMapReduceIndexMappedCount(stats *stats) float64 {
	value, _ := jp.GetFloat(stats.metrics, "MapReduceIndexes", "MappedPerSec", "Count")
	return value
}

func getMapReduceIndexReducedCount(stats *stats) float64 {
	value, _ := jp.GetFloat(stats.metrics, "MapReduceIndexes", "ReducedPerSec", "Count")
	return value
}

func getDatabaseDocumentCount(dbStats *dbStats) float64 {
	value, _ := jp.GetFloat(dbStats.databaseStats, "CountOfDocuments")
	return value
}

func createGauge(name string, help string) prometheus.Gauge {
	return prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
	})
}

func createGaugeDatabaseVec(name string, help string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
	}, []string { "database" })
}

func createCounter(name string, help string) prometheus.Counter {
	return prometheus.NewCounter(prometheus.CounterOpts {
		Namespace: 	namespace,
		Subsystem: 	subsystem,
		Name:		name,
		Help:		help,
	})
}

var timespanRegex = regexp.MustCompile(`((?P<days>\d+)\.)?(?P<hours>\d{2}):(?P<minutes>\d{2}):(?P<seconds>\d{2})(\.(?P<secondfraction>\d{7}))?`)

func timeSpanToSeconds(timespanString string) float64 {
	
	var result float64

	matches := matchNamedGroups(timespanRegex, timespanString)
	if daysString, ok := matches["days"]; ok {
		days, _ := strconv.Atoi(daysString)
		result = result + float64(days) * 24 * 60 * 60
	}
	if hoursString, ok := matches["hours"]; ok {
		hours, _ := strconv.Atoi(hoursString)
		result = result + float64(hours) * 60 * 60
	}
	if minutesString, ok := matches["minutes"]; ok {
		minutes, _ := strconv.Atoi(minutesString)
		result = result + float64(minutes) * 60
	}
	if secondsString, ok := matches["seconds"]; ok {
		seconds, _ := strconv.Atoi(secondsString)
		result = result + float64(seconds) 
	}
	if secondFractionString, ok := matches["secondfraction"]; ok {
		secondFraction, _ := strconv.Atoi(secondFractionString)
		result = result + float64(secondFraction) / 10000000
	}

	return result
}

func matchNamedGroups(regex *regexp.Regexp, text string) map[string]string {
	matches := regex.FindStringSubmatch(text)

	results := make(map[string]string)
	for i, name := range regex.SubexpNames() {
		if name != "" {
			results[name] = matches[i]
		}
	}
	return results
}
