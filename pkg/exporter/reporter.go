package exporter

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/dmartinol/application-exporter/pkg/config"
	logger "github.com/dmartinol/application-exporter/pkg/log"
)

type Reporter interface {
	Report(w *io.Writer)
}

type FileReporter struct {
	config       *config.Config
	runnerConfig *config.RunnerConfig
}

func NewFileReporter(config *config.Config, runnerConfig *config.RunnerConfig) FileReporter {
	return FileReporter{config: config, runnerConfig: runnerConfig}
}

func (r *FileReporter) Report(data *strings.Builder) {
	file, err := os.Create(fmt.Sprintf("%s.%s", r.runnerConfig.OutputFileName(), r.config.ContentType().Suffix()))
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
	config       *config.Config
	runnerConfig *config.RunnerConfig
	rw           http.ResponseWriter
}

func NewHttpReporter(config *config.Config, runnerConfig *config.RunnerConfig, rw http.ResponseWriter) HttpReporter {
	return HttpReporter{config: config, runnerConfig: runnerConfig, rw: rw}
}

func (r *HttpReporter) Report(data *strings.Builder) {
	// r.rw.WriteHeader(http.StatusOK)

	r.rw.Header().Set("Content-Type", r.config.ContentType().HttpContentType())
	r.rw.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s.%s", r.runnerConfig.OutputFileName(), r.config.ContentType().Suffix()))
	r.rw.Write([]byte(data.String()))
}
