package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/namsral/flag"

	"github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	namespace = "ravendb"
	subsystem = "exporter"
)

var (
	log = logrus.New()

	timeout time.Duration
	port    uint
	verbose bool

	ravenDbURL     string
	caCertFile     string
	useAuth        bool
	clientCertFile string
	clientKeyFile  string
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
		<head><title>RavenDB exporter for Prometheus</title></head>
		<body>
		<h1>RavenDB exporter for Prometheus</h1>
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

func readAndValidateConfig() {
	flag.StringVar(&ravenDbURL, "ravendb-url", "http://localhost:8080", "RavenDB URL")
	flag.UintVar(&port, "port", 9999, "Port to listen on")
	flag.DurationVar(&timeout, "timeout", time.Second*10, "Timeout when calling RavenDB")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")

	flag.StringVar(&caCertFile, "ca-cert", "", "Path to CA public cert file")
	flag.BoolVar(&useAuth, "use-auth", false, "If set, connection to RavenDB will be authenticated with a client certificate")
	flag.StringVar(&clientCertFile, "client-cert", "", "Path to client public certificate used for authentication")
	flag.StringVar(&clientKeyFile, "client-key", "", "Path to client private key used for authentication")

	flag.Parse()

	log.WithFields(logrus.Fields{
		"ravenDbUrl": ravenDbURL,
		"caCert":     caCertFile,
		"useAuth":    useAuth,
		"clientCert": clientCertFile,
		"clientKey":  clientKeyFile,
		"port":       port,
		"timeout":    timeout,
		"verbose":    verbose,
	}).Infof("RavenDB exporter configured")

	if useAuth && (caCertFile == "" || clientCertFile == "" || clientKeyFile == "") {
		log.Fatal("Invalid configuration: when using authentication you need to specify the CA cert, client cert and client private key")
	}
}

func setupLogger() {
	if verbose {
		log.Level = logrus.DebugLevel
	}
}

func startHTTPServer() {
	listenAddr := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

func main() {

	readAndValidateConfig()
	setupLogger()

	initializeClient()

	serveLandingPage()
	serveMetrics()

	startHTTPServer()
}
