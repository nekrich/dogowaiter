ARG GO_VERSION=1.26.0
ARG ALPINE_VERSION=latest

# Build stage. Build from repo root: docker build -f build/package/Dockerfile .
# TARGETOS/TARGETARCH are set by Docker Buildx when using --platform (e.g. linux/amd64, linux/arm64).
FROM golang:${GO_VERSION}-alpine AS build
ARG TARGETOS=linux
ARG TARGETARCH
WORKDIR /build
ENV CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH}
COPY . .
RUN go mod download
RUN mise-tasks/go/build

# Run stage
FROM alpine:${ALPINE_VERSION} AS container
COPY --from=build /build/dogowaiter /dogowaiter
ENTRYPOINT ["/dogowaiter"]
HEALTHCHECK --interval=10s --timeout=5s --start-period=1s --retries=5 \
  CMD ["/bin/sh", "-c", "grep -q '\"healthy\": true' \"${DOGOWAITER_HEALTH_FILE:-/tmp/healthy}\" || exit 1"]

ARG ALPINE_VERSION=latest
ARG IMAGE_VERSION=1.0.0
ARG IMAGE_REVISION=unknown
ARG IMAGE_CREATED_AT=unknown

# opencontainers image labels
LABEL \
  org.opencontainers.image.source="https://github.com/nekrich/dogowaiter" \
  org.opencontainers.image.title="dogowaiter" \
  org.opencontainers.image.description="Docker cross-stack dependency waiter in Go" \
  org.opencontainers.image.licenses="MIT" \
  org.opencontainers.image.url="https://github.com/nekrich/dogowaiter" \
  org.opencontainers.image.documentation="https://github.com/nekrich/dogowaiter/blob/main/README.md" \
  org.opencontainers.image.vendor="Vitalii Budnik" \
  org.opencontainers.image.authors="Vitalii Budnik <inekrich@me.com>" \
  org.opencontainers.image.version=${IMAGE_VERSION} \
  org.opencontainers.image.created=${IMAGE_CREATED_AT} \
  org.opencontainers.image.revision=${IMAGE_REVISION} \
  org.opencontainers.image.base.name=hub.docker.com/_/alpine:${ALPINE_VERSION}
