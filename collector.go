package main

import (
	jp "github.com/buger/jsonparser"
	"github.com/prometheus/client_golang/prometheus"
	"regexp"
	"strconv"
)

const (
	namespace = "ravendb"
	subsystem = ""
)

type exporter struct {
	up                         prometheus.Gauge
	workingSet                 prometheus.Gauge
	cpuTime                    prometheus.Counter
	isLeader                   prometheus.Gauge
	requestCount               prometheus.Counter
	documentPutCount           prometheus.Counter
	documentPutBytes           prometheus.Counter
	mapIndexIndexedCount       prometheus.Counter
	mapReduceIndexMappedCount  prometheus.Counter
	mapReduceIndexReducedCount prometheus.Counter

	databaseDocumentCount   *prometheus.GaugeVec
	databaseIndexCount      *prometheus.GaugeVec
	databaseStaleIndexCount *prometheus.GaugeVec
	databaseSize            *prometheus.GaugeVec

	databaseRequestCount               *prometheus.CounterVec
	databaseDocumentPutCount           *prometheus.CounterVec
	databaseDocumentPutBytes           *prometheus.CounterVec
	databaseMapIndexIndexedCount       *prometheus.CounterVec
	databaseMapReduceIndexMappedCount  *prometheus.CounterVec
	databaseMapReduceIndexReducedCount *prometheus.CounterVec
}

func newExporter() *exporter {
	return &exporter{
		up:                         		createGauge("up", "Whether the RavenDB scrape was successful"),
		workingSet:                 		createGauge("working_set_bytes", "Process working set"),
		cpuTime:                    		createCounter("cpu_time_seconds", "CPU time"),
		isLeader:                   		createGauge("is_leader", "If 1, then node is the cluster leader, otherwise 0"),
		requestCount:               		createCounter("request_count", "Server-wide request count"),
		documentPutCount:           		createCounter("document_put_count", "Server-wide document puts count"),
		documentPutBytes:           		createCounter("document_put_bytes", "Server-wide document put bytes"),
		mapIndexIndexedCount:       		createCounter("mapindex_indexed_count", "Server-wide map index indexed count"),
		mapReduceIndexMappedCount:  		createCounter("mapreduceindex_mapped_count", "Server-wide map-reduce index mapped count"),
		mapReduceIndexReducedCount: 		createCounter("mapreduceindex_reduced_count", "Server-wide map-reduce index reduced count"),

		databaseDocumentCount: 				createDatabaseGaugeVec("database_document_count", "Count of documents in a database"),
		databaseIndexCount:      			createDatabaseGaugeVec("database_index_count", "Count of indexes in a database"),
		databaseStaleIndexCount: 			createDatabaseGaugeVec("database_stale_index_cont", "Count of stale indexes in a database"),
		databaseSize:            			createDatabaseGaugeVec("database_size_bytes", "Database size in bytes"),
	
		databaseRequestCount:               createDatabaseCounterVec("database_request_count", "Database request count"),
		databaseDocumentPutCount:           createDatabaseCounterVec("database_document_put_count", "Database document puts count"),
		databaseDocumentPutBytes:           createDatabaseCounterVec("database_document_put_bytes", "Database document put bytes"),
		databaseMapIndexIndexedCount:       createDatabaseCounterVec("database_mapindex_indexed_count", "Database map index indexed count"),
		databaseMapReduceIndexMappedCount:  createDatabaseCounterVec("database_mapreduceindex_mapped_count", "Database map-reduce index mapped count"),
		databaseMapReduceIndexReducedCount: createDatabaseCounterVec("database_mapreduceindex_reduced_count", "Database map-reduce index reduced count"),
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

		collectPerDatabaseGauge(stats, e.databaseDocumentCount, getDatabaseDocumentCount, ch)
		collectPerDatabaseGauge(stats, e.databaseIndexCount, getDatabaseIndexCount, ch)
		collectPerDatabaseGauge(stats, e.databaseStaleIndexCount, getDatabaseStaleIndexCount, ch)
		collectPerDatabaseGauge(stats, e.databaseSize, getDatabaseSize, ch)

		collectPerDatabaseCounter(stats, e.databaseRequestCount, getDatabaseRequestCount, ch)
		collectPerDatabaseCounter(stats, e.databaseDocumentPutBytes, getDatabaseDocumentPutBytes, ch)
		collectPerDatabaseCounter(stats, e.databaseDocumentPutCount, getDatabaseDocumentPutCount, ch)

		collectPerDatabaseCounter(stats, e.databaseMapIndexIndexedCount, getDatabaseMapIndexIndexedCount, ch)
		collectPerDatabaseCounter(stats, e.databaseMapReduceIndexMappedCount, getDatabaseMapReduceIndexMappedCount, ch)
		collectPerDatabaseCounter(stats, e.databaseMapReduceIndexReducedCount, getDatabaseMapReduceIndexReducedCount, ch)
		
	}
}

func collectPerDatabaseGauge(stats *stats, vec *prometheus.GaugeVec, collectFunc func(*dbStats) float64, ch chan<- prometheus.Metric) {
	for _, dbs := range stats.dbStats {
		vec.WithLabelValues(dbs.database).Set(collectFunc(dbs))
	}
	vec.Collect(ch)
}

func collectPerDatabaseCounter(stats *stats, vec *prometheus.CounterVec, collectFunc func(*dbStats) float64, ch chan<- prometheus.Metric) {
	for _, dbs := range stats.dbStats {
		vec.WithLabelValues(dbs.database).Set(collectFunc(dbs))
	}
	vec.Collect(ch)
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

func getIsLeader(stats *stats) float64 {
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

func getDatabaseIndexCount(dbStats *dbStats) float64 {
	value, _ := jp.GetFloat(dbStats.databaseStats, "CountOfIndexes")
	return value
}

func getDatabaseStaleIndexCount(dbStats *dbStats) float64 {
	count := 0
	jp.ArrayEach(dbStats.databaseStats, func(value []byte, dataType jp.ValueType, offset int, err error) {
		if isStale, _ := jp.GetBoolean(value, "IsStale"); isStale {
			count++
		}
	}, "Indexes")
	
	return float64(count)
}

func getDatabaseSize(dbStats *dbStats) float64 {
	value, _ := jp.GetFloat(dbStats.databaseStats, "SizeOnDisk", "SizeInBytes")
	return value
}

func getDatabaseRequestCount(dbStats *dbStats) float64 {
	value, _ := jp.GetFloat(dbStats.metrics, "Requests", "RequestsPerSec", "Count")
	return value
}

func getDatabaseDocumentPutCount(dbStats *dbStats) float64 {
	value, _ := jp.GetFloat(dbStats.metrics, "Docs", "PutsPerSec", "Count")
	return value
}

func getDatabaseDocumentPutBytes(dbStats *dbStats) float64 {
	value, _ := jp.GetFloat(dbStats.metrics, "Docs", "BytesPutsPerSec", "Count")
	return value
}

func getDatabaseMapIndexIndexedCount(dbStats *dbStats) float64 {
	value, _ := jp.GetFloat(dbStats.metrics, "MapIndexes", "IndexedPerSec", "Count")
	return value
}

func getDatabaseMapReduceIndexMappedCount(dbStats *dbStats) float64 {
	value, _ := jp.GetFloat(dbStats.metrics, "MapIndexes", "MappedPerSec", "Count")
	return value
}

func getDatabaseMapReduceIndexReducedCount(dbStats *dbStats) float64 {
	value, _ := jp.GetFloat(dbStats.metrics, "MapIndexes", "ReducedPerSec", "Count")
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

func createDatabaseGaugeVec(name string, help string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
	}, []string{"database"})
}

func createCounter(name string, help string) prometheus.Counter {
	return prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
	})
}

func createDatabaseCounterVec(name string, help string) *prometheus.CounterVec {
	return prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
	}, []string{"database"})
}

var timespanRegex = regexp.MustCompile(`((?P<days>\d+)\.)?(?P<hours>\d{2}):(?P<minutes>\d{2}):(?P<seconds>\d{2})(\.(?P<secondfraction>\d{7}))?`)

func timeSpanToSeconds(timespanString string) float64 {

	var result float64

	matches := matchNamedGroups(timespanRegex, timespanString)
	if daysString, ok := matches["days"]; ok {
		days, _ := strconv.Atoi(daysString)
		result = result + float64(days)*24*60*60
	}
	if hoursString, ok := matches["hours"]; ok {
		hours, _ := strconv.Atoi(hoursString)
		result = result + float64(hours)*60*60
	}
	if minutesString, ok := matches["minutes"]; ok {
		minutes, _ := strconv.Atoi(minutesString)
		result = result + float64(minutes)*60
	}
	if secondsString, ok := matches["seconds"]; ok {
		seconds, _ := strconv.Atoi(secondsString)
		result = result + float64(seconds)
	}
	if secondFractionString, ok := matches["secondfraction"]; ok {
		secondFraction, _ := strconv.Atoi(secondFractionString)
		result = result + float64(secondFraction)/10000000
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
