FROM registry.access.redhat.com/ubi8/go-toolset:latest as builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY pkg/ ./pkg/
COPY pkg/ ./pkg/
COPY pkg/ ./pkg/

# RUN CGO_ENABLED=0 GOOS=linux go build -v -o /exporter exporter.go
# RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly  -v -o /exporter
RUN ls -lh /tmp/
RUN go build -v -o /tmp/exporter exporter.go
RUN ls -lh /tmp/
FROM registry.access.redhat.com/ubi8/ubi-minimal AS runner
COPY --from=builder /tmp/exporter /go/bin/exporter
RUN ls -la /go/bin/exporter

ENV SERVERPORT=8080
EXPOSE ${SERVERPORT}
ENTRYPOINT ["/go/bin/exporter"]
