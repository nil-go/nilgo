# It builds a minimal multi-platform Docker image for a Go application using multi-stage builds.
#
# See https://docs.docker.com/engine/reference/builder.
# See https://docs.docker.com/develop/develop-images/multistage-build/.

FROM golang:1.23-alpine as build

WORKDIR /app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -a -trimpath -o /app/app ./cmd


FROM gcr.io/distroless/static:nonroot

COPY --from=build /app/app /app
ENTRYPOINT ["/app"]
