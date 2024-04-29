# It builds a minimal multi-platform Docker image for a Go application using multi-stage builds.
#
# See https://docs.docker.com/engine/reference/builder.
# See https://docs.docker.com/develop/develop-images/multistage-build/.
# See https://docs.docker.com/build/building/multi-platform/

ARG GO_VERSION=1.22
FROM golang:$GO_VERSION-alpine as build

ARG TARGETARCH
ARG TARGETOS
ARG APP

WORKDIR /app
COPY . .
RUN GOARCH=$TARGETARCH GOOS=$TARGETOS CGO_ENABLED=0 go build -C examples/$APP -ldflags="-s -w" -a -trimpath -o /app/app ./cmd


FROM gcr.io/distroless/static:nonroot

COPY --from=build /app/app /app
ENTRYPOINT ["/app"]
