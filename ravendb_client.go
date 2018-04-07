package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	client http.Client
)

type stats struct {
	memory map[string]interface{}
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
	data, err := get("/admin/debug/memory/stats")
	if err != nil {
		return nil, err
	}

	return &stats{memory: data}, nil
}

func get(path string) (map[string]interface{}, error) {
	response, err := client.Get(ravenDbURL + path)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	buf, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err = json.Unmarshal(buf, &result); err != nil {
		return nil, err
	}

	return result, nil
}
