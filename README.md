# deployment-exporter
Go application to export the configuration of an OpenShift deployment


**WORK IN PROGRESS**

```bash
SERVER_PORT=8181 NS_SELECTOR=infinitySubChart=true LOG_LEVEL=debug go run main.go
```


```bash
curl "http://localhost:8181/inventory?type=text"
curl "http://localhost:8181/inventory?type=CSV"
```

RUNTIME_MODE=REST go run main.go
RUNTIME_MODE=app go run main.go

Automatic file suffix
```bash
go run main.go --help
  -as-service
        Run as REST service
  -content-type string
        Content type, one of text, CSV, JSON or YAML (default "text")
  -in-container
        Run in Container
  -log-level string
        Log level, one of debug, info, warn (default "info")
  -ns-selector string
        Namespace selector, like label1=value1,label2=value2
  -output string
        Output file name, default is output.<content-type> (unmanaged when runs as REST service). File suffix is automatically added
  -server-port int
        Server port (only for REST service mode) (default 8080)```

Following environment variables override the command arguments:
```yaml
AS_SERVICE (any value)
IN_CONTAINER (any value)
LOG_LEVEL
NS_SELECTOR
CONTENT_TYPE
SERVER_PORT
```

In REST mode, `content-type` query argument can override the `-content-type` and `CONTENT_TYPE` settings, as in:
`http://127.0.0.1:8181/inventory?content-type=CSV`

|Run as | Run in  | Formatter |
|--- | --- | ---|
|Script|Container|Dump on file (*)|
|Script|Local|Dump on file|
|REST|Container, Local|Output on HTTP response|

|Run as | Run in  | Formatter |
|--- | --- | ---|
|Script|Container|Dump on file (*)|
|Script|Local|Dump on file|
|REST|Container, Local|Output on HTTP response|

(*) The container exits after the application completes, you need to add sleeps in order to collect the output file