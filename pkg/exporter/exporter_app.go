package exporter

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/dmartinol/application-exporter/pkg/config"
	cfg "github.com/dmartinol/application-exporter/pkg/config"
	"github.com/dmartinol/application-exporter/pkg/formatter"
	logger "github.com/dmartinol/application-exporter/pkg/log"
	"github.com/dmartinol/application-exporter/pkg/model"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ExporterApp struct {
	config       *cfg.Config
	runnerConfig *cfg.RunnerConfig
}

func NewExporterApp(config *cfg.Config) *ExporterApp {
	return &ExporterApp{config: config, runnerConfig: config.GlobalRunnerConfig()}
}

func (app *ExporterApp) Start() {
	runner := app.newRunner()
	RunExporter(runner, app.runnerConfig)
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

func (r ExporterAppRunner) Collect(runnerConfig *cfg.RunnerConfig, kubeConfig *rest.Config) (*model.TopologyModel, error) {
	topology, err := NewModelBuilder(r.config, runnerConfig).BuildForKubeConfig(kubeConfig)
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
func (r ExporterAppRunner) Report(runnerConfig *cfg.RunnerConfig, output *strings.Builder) {
	reporter := NewFileReporter(r.config, runnerConfig)
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
