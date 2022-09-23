package exporter

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	logger "github.com/dmartinol/application-exporter/pkg/log"
)

type Reporter interface {
	Report(w *io.Writer)
}

type FileReporter struct {
	config *Config
}

func NewFileReporter(config *Config) FileReporter {
	return FileReporter{config: config}
}

func (r *FileReporter) Report(data *strings.Builder) {
	file, err := os.Create(fmt.Sprintf("%s.%s", r.config.OutputFileName(), r.config.contentType.Suffix()))
	logger.Infof("Printing output on %s", file.Name())
	if err != nil {
		log.Fatalln("Error creating output file", err)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	defer w.Flush()

	w.WriteString(data.String())
}

type HttpReporter struct {
	config *Config
	rw     http.ResponseWriter
}

func NewHttpReporter(config *Config, rw http.ResponseWriter) HttpReporter {
	return HttpReporter{config: config, rw: rw}
}

func (r *HttpReporter) Report(data *strings.Builder) {
	// r.rw.WriteHeader(http.StatusOK)

	r.rw.Header().Set("Content-Type", r.config.contentType.HttpContentType())
	r.rw.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s.%s", r.config.OutputFileName(), r.config.contentType.Suffix()))
	r.rw.Write([]byte(data.String()))
}
