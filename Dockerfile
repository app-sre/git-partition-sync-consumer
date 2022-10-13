FROM quay.io/app-sre/golang:1.18.5 as builder
WORKDIR /build
COPY . .
RUN make build

FROM registry.access.redhat.com/ubi8-minimal
COPY --from=builder /build/gitlab-sync-pull  /bin/gitlab-sync-pull
RUN microdnf install git

ENTRYPOINT  [ "/bin/gitlab-sync-pull" ]