FROM golang:1.14-alpine as builder

RUN mkdir /build
WORKDIR /build

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o ./go-hackernews ./cmd/web

FROM scratch
COPY --from=builder /build/go-hackernews /app/
COPY ./ui /app/ui
COPY ./config.yml /app/
WORKDIR /app

# This container exposes port 5000 to the outside world
EXPOSE 5000

# Run the binary program produced by `go build`
CMD ["./go-hackernews"]