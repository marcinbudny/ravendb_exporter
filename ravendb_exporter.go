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

var (
	log = logrus.New()

	timeout time.Duration
	port    uint
	verbose bool

	ravenDbURL        string
	caCertFile        string
	useAuth           bool
	clientCertFile    string
	clientKeyFile     string
	clientKeyPassword string
)

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
	prometheus.MustRegister(newExporter())

	http.Handle("/metrics", promhttp.Handler())
}

func readAndValidateConfig() {
	flag.String(flag.DefaultConfigFlagname, "", "path to config file")
	flag.StringVar(&ravenDbURL, "ravendb-url", "http://localhost:8080", "RavenDB URL")
	flag.UintVar(&port, "port", 9440, "Port to expose scraping endpoint on")
	flag.DurationVar(&timeout, "timeout", time.Second*10, "Timeout when calling RavenDB")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")

	flag.StringVar(&caCertFile, "ca-cert", "", "Path to CA public cert file of RavenDB server")
	flag.BoolVar(&useAuth, "use-auth", false, "If set, connection to RavenDB will be authenticated with a client certificate")
	flag.StringVar(&clientCertFile, "client-cert", "", "Path to client public certificate used for authentication")
	flag.StringVar(&clientKeyFile, "client-key", "", "Path to client private key used for authentication")
	flag.StringVar(&clientKeyPassword, "client-key-password", "", "(optional) Password for the client private keys")

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
