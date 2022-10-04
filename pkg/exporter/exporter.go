package exporter

import (
	"strings"

	logger "github.com/dmartinol/application-exporter/pkg/log"
	"github.com/dmartinol/application-exporter/pkg/model"
	"k8s.io/client-go/rest"
)

type Exporter interface {
	Start()
}

type ExporterRunner interface {
	Connect() (*rest.Config, error)
	Collect(*rest.Config) (*model.TopologyModel, error)
	Transform(topology *model.TopologyModel) *strings.Builder
	Report(output *strings.Builder)
}

func RunExporter(runner ExporterRunner) error {
	kubeConfig, err := runner.Connect()
	if err != nil {
		logger.Fatalf("Cannot connect cluster: %s", err)
		return err
	}

	logger.Info("Cluster connected")
	topology, err := runner.Collect(kubeConfig)
	if err != nil {
		return err
	}

	output := runner.Transform(topology)
	runner.Report(output)

	return nil
}
