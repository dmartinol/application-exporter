package exporter

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	logger "github.com/dmartinol/application-exporter/pkg/log"
	"github.com/dmartinol/application-exporter/pkg/model"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gorilla/mux"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeconfig *string
var router = mux.NewRouter()

type ExporterService struct {
	config *Config
}

func NewExporterService(config *Config) *ExporterService {
	return &ExporterService{config: config}
}

func (s *ExporterService) Start() {
	NewExporterMetrics(s.config)
	logger.Infof("Started metrics")

	router.Path("/inventory").Queries("content-type", "{content-type}").Queries("ns-selector", "{ns-selector}").Queries("output", "{output}").Queries("with-resources", "{with-resources}").HandlerFunc(s.inventoryHandler).Name("inventoryHandler")
	router.Path("/inventory").HandlerFunc(s.inventoryHandler).Name("inventoryHandler")
	router.Path("/metrics").Handler(promhttp.Handler())

	host := "localhost"
	if s.config.RunInContainer() {
		host = "0.0.0.0"
	}
	url := fmt.Sprintf("%s:%d", host, s.config.ServerPort())
	logger.Infof("Starting listener as %s", url)
	if err := http.ListenAndServe(url, router); err != nil {
		logger.Fatal(err)
	}

}

type ExporterServiceRunner struct {
	config *Config
	rw     http.ResponseWriter
	req    *http.Request
}

func (s *ExporterService) newRunner(config *Config, rw http.ResponseWriter, req *http.Request) ExporterServiceRunner {
	runner := ExporterServiceRunner{}
	runner.rw = rw
	runner.req = req
	runner.config = config

	return runner
}

func (s *ExporterService) inventoryHandler(rw http.ResponseWriter, req *http.Request) {
	newConfig := *s.config

	contentTypeArg := req.FormValue("content-type")
	if contentTypeArg != "" {
		newConfig.contentType = ContentTypeFromString(contentTypeArg)
	}
	namespaceSelector := req.FormValue("ns-selector")
	if namespaceSelector != "" {
		newConfig.namespaceSelector = namespaceSelector
	}
	outputFileName := req.FormValue("output")
	if outputFileName != "" {
		newConfig.outputFileName = outputFileName
	}
	withResources := req.FormValue("with-resources")
	if withResources != "" {
		newConfig.withResources = true
	}
	burstArg := req.FormValue("burst")
	if burstArg != "" {
		burst, err := strconv.Atoi(burstArg)
		if err != nil {
			logger.Warnf("Disregarding non numeric value %s", req.FormValue("burst"))
		} else {
			newConfig.burst = burst
		}
	}

	if req.URL.Path == "/inventory" {
		if req.Method == "POST" {
			runner := s.newRunner(&newConfig, rw, req)
			RunExporter(runner)
		} else {
			http.Error(rw, fmt.Sprintf("Expect method POST at /, got %v", req.Method), http.StatusMethodNotAllowed)
		}
	} else {
		http.Error(rw, fmt.Sprintf("Unmanaged path %s", req.URL.Path), http.StatusNotFound)
		return
	}
}

func (r ExporterServiceRunner) Connect() (*rest.Config, error) {
	kubeConfig, err := r.connectCluster()
	if err != nil {
		http.Error(r.rw, fmt.Sprintf("Cannot connect cluster: %s", err), http.StatusInternalServerError)
	}
	return kubeConfig, err
}

func (r ExporterServiceRunner) Collect(kubeConfig *rest.Config) (*model.TopologyModel, error) {
	topology, err := NewModelBuilder(r.config).BuildForKubeConfig(kubeConfig)
	if err != nil {
		http.Error(r.rw, fmt.Sprintf("Cannot build data model: %s", err), http.StatusInternalServerError)
		return nil, err
	}
	return topology, nil
}

func (r ExporterServiceRunner) Transform(topology *model.TopologyModel) *strings.Builder {
	fmt := NewFormatterForConfig(r.config)
	return fmt.Format(topology)
}

func (r ExporterServiceRunner) Report(output *strings.Builder) {
	reporter := NewHttpReporter(r.config, r.rw)
	reporter.Report(output)
}

func (s ExporterServiceRunner) initKubeconfig() *string {
	if home := s.homeDir(); home != "" {
		return flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "")
	} else {
		return flag.String("kubeconfig", "", "")
	}
}

func (s ExporterServiceRunner) connectCluster() (*rest.Config, error) {
	if s.config.RunInContainer() {
		return rest.InClusterConfig()
	} else {
		if kubeconfig == nil {
			kubeconfig = s.initKubeconfig()
		}
		//Load config for Openshift's go-client from kubeconfig file
		return clientcmd.BuildConfigFromFlags("", *kubeconfig)
	}
}

func (s ExporterServiceRunner) homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
