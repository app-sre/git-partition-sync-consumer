FROM registry.access.redhat.com/ubi8/go-toolset:1.21.13-1.1727869850 as builder
LABEL konflux.additional-tags=1.0.0
USER 0
WORKDIR /workspace
COPY . .
RUN make build

FROM builder AS test
RUN make test

FROM registry.access.redhat.com/ubi9-minimal:9.6-1754000177 as prod
COPY --from=builder /workspace/git-partition-sync-consumer  /bin/git-partition-sync-consumer
RUN microdnf update -y && microdnf install -y git

ENTRYPOINT  [ "/bin/git-partition-sync-consumer" ]
