FROM quay.io/app-sre/golang:1.18.5 AS builder
WORKDIR /build
COPY . .
RUN make test build

FROM registry.access.redhat.com/ubi9-minimal:9.4
COPY --from=builder /build/git-partition-sync-consumer  /bin/git-partition-sync-consumer
RUN microdnf update -y && microdnf install -y git

ENTRYPOINT  [ "/bin/git-partition-sync-consumer" ]
