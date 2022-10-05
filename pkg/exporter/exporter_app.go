package exporter

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/dmartinol/application-exporter/pkg/config"
	"github.com/dmartinol/application-exporter/pkg/formatter"
	logger "github.com/dmartinol/application-exporter/pkg/log"
	"github.com/dmartinol/application-exporter/pkg/model"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ExporterApp struct {
	config *config.Config
}

func NewExporterApp(config *config.Config) *ExporterApp {
	return &ExporterApp{config: config}
}

func (app *ExporterApp) Start() {
	runner := app.newRunner()
	RunExporter(runner)
}

type ExporterAppRunner struct {
	config *config.Config
}

func (app *ExporterApp) newRunner() ExporterAppRunner {
	runner := ExporterAppRunner{}
	runner.config = app.config

	return runner
}

func (r ExporterAppRunner) Connect() (*rest.Config, error) {
	//Load config for Openshift's go-client from kubeconfig file
	return clientcmd.BuildConfigFromFlags("", *r.initKubeconfig())
}

func (r ExporterAppRunner) Collect(kubeConfig *rest.Config) (*model.TopologyModel, error) {
	topology, err := NewModelBuilder(r.config).BuildForKubeConfig(kubeConfig)
	if err != nil {
		logger.Fatalf("Cannot build data model", err)
		return nil, err
	}
	return topology, nil
}

func (r ExporterAppRunner) Transform(topology *model.TopologyModel) *strings.Builder {
	fmt := formatter.NewFormatterForConfig(r.config)
	output := fmt.Format(topology)
	return output
}
func (r ExporterAppRunner) Report(output *strings.Builder) {
	reporter := NewFileReporter(r.config)
	reporter.Report(output)
}

func (r ExporterAppRunner) initKubeconfig() *string {
	if home := r.homeDir(); home != "" {
		return flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "")
	} else {
		return flag.String("kubeconfig", "", "")
	}
}

func (r ExporterAppRunner) homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
