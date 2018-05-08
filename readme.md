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

## Configuration

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

Sample configuration with authentication, for Docker:

```bash
docker run -d \
-e RAVENDB_URL=https://a.myserver.ravendb.community \
-e CA_CERT=/certs/lets-encrypt-x3-cross-signed.crt \
-e USE_AUTH=true \
-e CLIENT_CERT=/certs/admin.client.certificate.myserver.crt \
-e CLIENT_KEY=/certs/admin.client.certificate.myserver.key \
-v /path/to/certs/on/host:/certs \
-p 9440:9440 \
marcinbudny/ravendb_exporter
```

## Exported metrics

Let me know if there is a metric you would like to be added.

### Server-wide

* **ravendb_up** - Whether the RavenDB scrape was successful
* **ravendb_working_set_bytes** - Process working set
* **ravendb_cpu_time_seconds** - CPU time
* **ravendb_is_leader** - If 1, then node is the cluster leader, otherwise 0
* **ravendb_request_count** - Server-wide request count
* **ravendb_document_put_count** - Server-wide document puts count
* **ravendb_document_put_bytes** - Server-wide document put bytes
* **ravendb_mapindex_indexed_count** - Server-wide map index indexed count
* **ravendb_mapreduceindex_mapped_count** - Server-wide map-reduce index mapped count
* **ravendb_mapreduceindex_reduced_count** - Server-wide map-reduce index reduced count

### Per database

* **ravendb_database_document_count** - Count of documents in a database
* **ravendb_database_index_count** - Count of indexes in a database
* **ravendb_database_stale_index_count** - Count of stale indexes in a database
* **ravendb_database_size_bytes** - Database size in bytes
* **ravendb_database_request_count** - Database request count
* **ravendb_database_document_put_count** - Database document puts count
* **ravendb_database_document_put_bytes** - Database document put bytes
* **ravendb_database_mapindex_indexed_count** - Database map index indexed count
* **ravendb_database_mapreduceindex_mapped_count** - Database map-reduce index  mapped count
* **ravendb_database_mapreduceindex_reduced_count** - Database map-reduce index reduced count


## Changelog

### 0.1.2

* Fixed name of the ravendb_database_stale_index_count metric

### 0.1.1

* Initial version