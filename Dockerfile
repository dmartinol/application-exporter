FROM ubi8/go-toolset:latest as builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY pkg/ ./pkg/

RUN ls -lh /tmp/
# RUN go build -v -o /tmp/exporter exporter.go
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly  -v -o /tmp/exporter
RUN ls -lh /tmp/
FROM ubi8/ubi-minimal AS runner
COPY --from=builder /tmp/exporter /go/bin/exporter
RUN ls -la /go/bin/exporter

ENV SERVERPORT=8080
EXPOSE ${SERVERPORT}
ENTRYPOINT ["/go/bin/exporter"]
