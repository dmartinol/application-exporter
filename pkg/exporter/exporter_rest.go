package exporter

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	logger "github.com/dmartinol/application-exporter/pkg/log"

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

func (s *ExporterService) Run() {
	router.Path("/inventory").Queries("content-type", "{content-type}").Queries("ns-selector", "{ns-selector}").Queries("output", "{output}").HandlerFunc(s.inventoryHandler).Name("inventoryHandler")
	router.Path("/inventory").HandlerFunc(s.inventoryHandler).Name("inventoryHandler")

	url := fmt.Sprintf("localhost:%d", s.config.ServerPort())
	logger.Infof("Starting listener as %s", url)
	if err := http.ListenAndServe(url, router); err != nil {
		logger.Fatal(err)
	}
}

func (s *ExporterService) inventoryHandler(rw http.ResponseWriter, req *http.Request) {
	newConfig := s.config

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

	if req.URL.Path == "/inventory" {
		if req.Method == "GET" {
			s.inventory(newConfig, rw)
		} else {
			http.Error(rw, fmt.Sprintf("Expect method GET at /, got %v", req.Method), http.StatusMethodNotAllowed)
		}
	} else {
		http.Error(rw, fmt.Sprintf("Unmanaged path %s", req.URL.Path), http.StatusNotFound)
		return
	}
}

func (s *ExporterService) inventory(newConfig *Config, rw http.ResponseWriter) {
	kubeConfig, err := s.connectCluster()
	if err != nil {
		http.Error(rw, fmt.Sprintf("Cannot connect cluster: %s", err), http.StatusInternalServerError)
	}
	logger.Info("Cluster connected")

	topology, err := NewModelBuilder(s.config).BuildForKubeConfig(kubeConfig)
	if err != nil {
		http.Error(rw, fmt.Sprintf("Cannot build data model: %s", err), http.StatusInternalServerError)
	}

	fmt := NewFormatterForConfig(newConfig)
	output := fmt.Format(topology)
	reporter := NewHttpReporter(s.config, rw)
	reporter.Report(output)
}

func (s *ExporterService) initKubeconfig() *string {
	if home := s.homeDir(); home != "" {
		return flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "")
	} else {
		return flag.String("kubeconfig", "", "")
	}
}

func (s *ExporterService) connectCluster() (*rest.Config, error) {
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

func (s *ExporterService) homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
