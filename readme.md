# RavenDB Prometheus exporter

Exports RavenDB metrics and allows for Prometheus scraping. Versions prior to 4 are not supported due to different API and authentication mechanism.

## Installation

### From source

You need to have a Go 1.20+ environment configured.

```bash
go install github.com/marcinbudny/ravendb_exporter@latest
ravendb_exporter --ravendb-url=http://live-test.ravendb.net
```

### Using Docker

```bash
docker run -d -p 9440:9440 -e RAVENDB_URL=http://live-test.ravendb.net marcinbudny/ravendb_exporter
```

## Configuration

The exporter can be configured with commandline arguments, environment variables and a configuration file. For the details on how to format the configuration file, visit [namsral/flag](https://github.com/namsral/flag) repo.

|Flag|ENV variable|Default|Meaning|
|---|---|---|---|
|--ravendb-url|RAVENDB_URL|http://localhost:8080|RavenDB URL|
|--port|PORT|9440|Port to expose scrape endpoint on|
|--timeout|TIMEOUT|10s|Timeout when calling RavenDB|
|--verbose|VERBOSE|false|Enable verbose logging|
|--ca-cert|CA_CERT|(empty)|Path to CA public cert file of RavenDB server|
|--use-auth|USE_AUTH|false|If set, connection to RavenDB will be authenticated with a client certificate|
|--client-cert|CLIENT_CERT|(empty)|Path to client public certificate used for authentication|
|--client-key|CLIENT_KEY|(empty)|Path to client private key used for authentication|
|--client-key-password|CLIENT_KEY_PASSWORD|(empty)|Password for the client key (if it is encrypted)|

Sample configuration with authentication, for Docker:

```bash
docker run -d \
-e RAVENDB_URL=https://a.myserver.ravendb.community \
-e CA_CERT=/certs/lets-encrypt-x3-cross-signed.crt \
-e USE_AUTH=true \
-e CLIENT_CERT=/certs/admin.client.certificate.myserver.crt \
-e CLIENT_KEY=/certs/admin.client.certificate.myserver.key \
-e CLIENT_KEY_PASSWORD=mypassword
-v /path/to/certs/on/host:/certs \
-p 9440:9440 \
marcinbudny/ravendb_exporter
```

## Exported metrics

Let me know if there is a metric you would like to be added.

```
ravendb_cpu_time_seconds_total 1613.68
# HELP ravendb_database_document_put_bytes_total Database document put bytes
# TYPE ravendb_database_document_put_bytes_total counter
ravendb_database_document_put_bytes_total{database="Demo"} 405
# HELP ravendb_database_document_put_total Database document puts count
# TYPE ravendb_database_document_put_total counter
ravendb_database_document_put_total{database="Demo"} 3
# HELP ravendb_database_documents Count of documents in a database
# TYPE ravendb_database_documents gauge
ravendb_database_documents{database="Demo"} 1063
# HELP ravendb_database_indexes Count of indexes in a database
# TYPE ravendb_database_indexes gauge
ravendb_database_indexes{database="Demo"} 20
# HELP ravendb_database_mapindex_indexed_total Database map index indexed count
# TYPE ravendb_database_mapindex_indexed_total counter
ravendb_database_mapindex_indexed_total{database="Demo"} 952
# HELP ravendb_database_mapreduceindex_mapped_total Database map-reduce index mapped count
# TYPE ravendb_database_mapreduceindex_mapped_total counter
ravendb_database_mapreduceindex_mapped_total{database="Demo"} 0
# HELP ravendb_database_mapreduceindex_reduced_total Database map-reduce index reduced count
# TYPE ravendb_database_mapreduceindex_reduced_total counter
ravendb_database_mapreduceindex_reduced_total{database="Demo"} 0
# HELP ravendb_database_request_total Database request count
# TYPE ravendb_database_request_total counter
ravendb_database_request_total{database="Demo"} 6179
# HELP ravendb_database_size_bytes Database size in bytes
# TYPE ravendb_database_size_bytes gauge
ravendb_database_size_bytes{database="Demo"} 6.35568128e+08
# HELP ravendb_database_stale_indexes Count of stale indexes in a database
# TYPE ravendb_database_stale_indexes gauge
ravendb_database_stale_indexes{database="Demo"} 0
# HELP ravendb_database_tasks Tasks in a database
# TYPE ravendb_database_tasks gauge
ravendb_database_tasks{connection_status="Active",database="Demo",type="Backup"} 2
ravendb_database_tasks{connection_status="NotActive",database="Demo",type="Backup"} 1
ravendb_database_tasks{connection_status="NotActive",database="Demo",type="Subscription"} 2
# HELP ravendb_document_put_bytes_total Server-wide document put bytes
# TYPE ravendb_document_put_bytes_total counter
ravendb_document_put_bytes_total 0
# HELP ravendb_document_put_total Server-wide document puts count
# TYPE ravendb_document_put_total counter
ravendb_document_put_total 0
# HELP ravendb_is_leader If 1, then node is the cluster leader, otherwise 0
# TYPE ravendb_is_leader gauge
ravendb_is_leader 1
# HELP ravendb_mapindex_indexed_total Server-wide map index indexed count
# TYPE ravendb_mapindex_indexed_total counter
ravendb_mapindex_indexed_total 0
# HELP ravendb_mapreduceindex_mapped_total Server-wide map-reduce index mapped count
# TYPE ravendb_mapreduceindex_mapped_total counter
ravendb_mapreduceindex_mapped_total 0
# HELP ravendb_mapreduceindex_reduced_total Server-wide map-reduce index reduced count
# TYPE ravendb_mapreduceindex_reduced_total counter
ravendb_mapreduceindex_reduced_total 0
# HELP ravendb_request_total Server-wide request count
# TYPE ravendb_request_total counter
ravendb_request_total 15530
# HELP ravendb_up Whether the RavenDB scrape was successful
# TYPE ravendb_up gauge
ravendb_up 1
# HELP ravendb_working_set_bytes Process working set
# TYPE ravendb_working_set_bytes gauge
ravendb_working_set_bytes 1.651195904e+09
```

## Changelog

### 0.5.1

* Fixed issue with per database metric not properly resetting

### 0.5.0

* Added database gauge metric: `ravendb_database_tasks`

### 0.4.0

* Updated golang version, some of the dependencies
* Used `scratch` as base for the docker image

### 0.3.0

* Support for password protected private keys
* Fixed problem with no option to load config file #1

### 0.2.0 

* Changed metric names to conform to Prometheus guidelines

### 0.1.2

* Fixed name of the ravendb_database_stale_index_count metric

### 0.1.1

* Initial version