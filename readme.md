# RavenDB 4 Prometheus exporter

Exports RavenDB 4 metrics and allows for Prometheus scraping. Versions prior to 4 are not supported due to different API and authentication mechanism.

## Installation

### From source

You need to have a Go 1.6+ environment configured.

```bash
cd $GOPATH/src
mkdir -p github.com/marcinbudny
git clone https://github.com/marcinbudny/ravendb_exporter github.com/marcinbudny/ravendb_exporter
cd github.com/marcinbudny/ravendb_exporter 
go build -o ravendb_exporter
./ravendb_exporter --ravendb-url=http://live-test.ravendb.net
```

### Using Docker

```bash
docker run -d -p 9440:9440 -e RAVENDB_URL=http://live-test.ravendb.net marcinbudny/ravendb_exporter
```

### Configuration

The exporter can be configured with commandline arguments, environment variables and a configuration file. For the details on how to format the configuration file, visit [namsral/flag](https://github.com/namsral/flag) repo.

|Commandline|ENV variable|Default|Meaning|
|---|---|---|---|
|--ravendb-url|RAVENDB_URL|http://localhost:8080|RavenDB URL|
|--port|PORT|9440|Port to expose scrape endpoint on|
|--timeout|TIMEOUT|10s|Timeout when calling RavenDB|
|--verbose|VERBOSE|false|Enable verbose logging|
|--ca-cert|CA_CERT|(empty)|Path to CA public cert file of RavenDB server|
|--use-auth|USE_AUTH|false|If set, connection to RavenDB will be authenticated with a client certificate|
|--client-cert|CLIENT_CERT|(empty)|Path to client public certificate used for authentication|
|--client-key|CLIENT_KEY|(empty)|Path to client private key used for authentication|

## Exported metrics

_TODO_

## Changelog

### 0.1.1

* Initial version