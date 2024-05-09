# It builds a minimal multi-platform Docker image for a Go application using multi-stage builds.
#
# See https://docs.docker.com/engine/reference/builder.
# See https://docs.docker.com/develop/develop-images/multistage-build/.

FROM golang:1.22 as build

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -a -trimpath -o /app/app ./cmd


FROM gcr.io/distroless/static:nonroot

COPY --from=build /app/app /app
ENTRYPOINT ["/app"]
