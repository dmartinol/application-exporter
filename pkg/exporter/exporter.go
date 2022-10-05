package exporter

import (
	"strings"

	"github.com/dmartinol/application-exporter/pkg/config"
	logger "github.com/dmartinol/application-exporter/pkg/log"
	"github.com/dmartinol/application-exporter/pkg/model"
	"k8s.io/client-go/rest"
)

type Exporter interface {
	Start()
}

type ExporterRunner interface {
	Connect() (*rest.Config, error)
	Collect(runnerConfig *config.RunnerConfig, config *rest.Config) (*model.TopologyModel, error)
	Transform(topology *model.TopologyModel) *strings.Builder
	Report(runnerConfig *config.RunnerConfig, output *strings.Builder)
}

func RunExporter(runner ExporterRunner, runnerConfig *config.RunnerConfig) error {
	kubeConfig, err := runner.Connect()
	if err != nil {
		logger.Fatalf("Cannot connect cluster: %s", err)
		return err
	}

	logger.Info("Cluster connected")
	topology, err := runner.Collect(runnerConfig, kubeConfig)
	if err != nil {
		return err
	}

	output := runner.Transform(topology)
	runner.Report(runnerConfig, output)

	return nil
}
