package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	jp "github.com/buger/jsonparser"
	"io/ioutil"
	"net/http"
)

var (
	client http.Client
)

type stats struct {
	cpu      []byte
	memory   []byte
	metrics  []byte
	nodeInfo []byte
	dbStats  []*dbStats
}

type dbStats struct {
	database        string
	collectionStats []byte
	metrics         []byte
	indexes         []byte
	databaseStats   []byte
	storage         []byte
}

func initializeClient() {
	tlsConfig := &tls.Config{}

	if caCertFile != "" {
		caCertData, err := ioutil.ReadFile(caCertFile)
		if err != nil {
			log.Fatal(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCertData)
		tlsConfig.RootCAs = caCertPool
	}

	if useAuth {
		cert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
		if err != nil {
			log.Fatal(err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
		tlsConfig.BuildNameToCertificate()
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}

	client = http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

}

func getStats() (*stats, error) {

	databases, err := getDatabaseNames()
	if err != nil {
		return nil, err
	}

	paths := preparePaths(databases)

	results := getAllPaths(paths, 16)

	return organizeGetResults(results, databases)
}

func getDatabaseNames() ([]string, error) {
	data, err := get("/databases")
	if err != nil {
		return nil, err
	}
	var databases []string

	dbsNode, _, _, _ := jp.Get(data, "Databases")
	jp.ArrayEach(dbsNode, func(value []byte, dataType jp.ValueType, offset int, err error) {
		database, _ := jp.GetString(value, "Name")
		databases = append(databases, database)
	})

	return databases, nil
}

func preparePaths(databases []string) []string {
	paths := []string{
		"/admin/debug/cpu/stats",
		"/admin/debug/memory/stats",
		"/admin/metrics",
		"/cluster/node-info",
	}

	for _, database := range databases {
		paths = append(paths, fmt.Sprintf("/databases/%s/collections/stats", database))
		paths = append(paths, fmt.Sprintf("/databases/%s/indexes", database))
		paths = append(paths, fmt.Sprintf("/databases/%s/metrics", database))
		paths = append(paths, fmt.Sprintf("/databases/%s/stats", database))
		paths = append(paths, fmt.Sprintf("/databases/%s/debug/storage/report", database))
	}

	return paths
}

func getAllPaths(paths []string, maxParallelism int) map[string]getResult {

	pathsChan := make(chan string)
	resultChan := make(chan getResult, len(paths))

	var doneChans []<-chan bool

	for i := 0; i < maxParallelism; i++ {
		doneChans = append(doneChans, getWorker(pathsChan, resultChan))
	}

	for _, path := range paths {
		pathsChan <- path
	}
	close(pathsChan)

	for i := 0; i < maxParallelism; i++ {
		<-doneChans[i]
	}

	allResults := make(map[string]getResult)
	for i := 0; i < len(paths); i++ {
		result := <-resultChan
		allResults[result.path] = result
	}

	return allResults
}

func getWorker(paths <-chan string, results chan<- getResult) <-chan bool {
	done := make(chan bool)

	go func() {
		for path := range paths {
			result, err := get(path)
			results <- getResult{path, result, err}
		}
		done <- true
	}()

	return done
}

type getResult struct {
	path   string
	result []byte
	err    error
}

func get(path string) ([]byte, error) {
	url := ravenDbURL + path

	log.WithField("url", url).Debug("GET request to RavenDB")

	response, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	buf, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func organizeGetResults(results map[string]getResult, databases []string) (*stats, error) {

	for _, result := range results {
		if result.err != nil {
			return nil, result.err
		}
	}

	stats := stats{
		cpu:      results["/admin/debug/cpu/stats"].result,
		memory:   results["/admin/debug/memory/stats"].result,
		metrics:  results["/admin/metrics"].result,
		nodeInfo: results["/cluster/node-info"].result,
	}

	for _, database := range databases {
		dbs := &dbStats{
			database:        database,
			collectionStats: results[fmt.Sprintf("/databases/%s/collections/stats", database)].result,
			indexes:         results[fmt.Sprintf("/databases/%s/indexes", database)].result,
			metrics:         results[fmt.Sprintf("/databases/%s/metrics", database)].result,
			databaseStats:   results[fmt.Sprintf("/databases/%s/stats", database)].result,
			storage:         results[fmt.Sprintf("/databases/%s//debug/storage/report", database)].result,
		}

		stats.dbStats = append(stats.dbStats, dbs)
	}

	return &stats, nil
}
