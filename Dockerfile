FROM golang:1.20.2-alpine as build

WORKDIR /go/src/github.com/marcinbudny/ravendb_exporter
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -o app

FROM scratch
COPY --from=build /go/src/github.com/marcinbudny/ravendb_exporter/app /
EXPOSE 9448
ENTRYPOINT [ "/app" ]