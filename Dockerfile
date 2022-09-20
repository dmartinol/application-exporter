FROM registry.access.redhat.com/ubi8/go-toolset:latest as builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY pkg/ ./pkg/

# RUN CGO_ENABLED=0 GOOS=linux go build -v -o /exporter exporter.go
# RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly  -v -o /exporter
RUN go build -v -o /tmp/exporter exporter.go

FROM registry.access.redhat.com/ubi8/ubi-minimal AS runner
COPY --from=builder /tmp/exporter /go/bin/exporter
EXPOSE 8080
ENV SERVERPORT=8080
ENTRYPOINT ["/go/bin/exporter"]
