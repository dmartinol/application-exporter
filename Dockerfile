FROM golang:latest as builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY pkg/ ./pkg/

RUN ls -lh /tmp/
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly  -v -o /tmp/exporter

FROM alpine:3 as runner
RUN apk add --no-cache ca-certificates
COPY --from=builder /tmp/exporter /go/bin/exporter

ENV SERVER_PORT=8080
ENV CONTAINER_MODE='true'
ENV LOG_LEVEL='info'
ENV NS_SELECTOR='label=value'
EXPOSE ${SERVER_PORT}
ENTRYPOINT ["/go/bin/exporter"]
