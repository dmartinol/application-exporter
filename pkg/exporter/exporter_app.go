package exporter

import (
	"flag"
	"os"
	"path/filepath"

	logger "github.com/dmartinol/deployment-exporter/pkg/log"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ExporterApp struct {
	config *Config
}

func NewExporterApp(config *Config) *ExporterApp {
	return &ExporterApp{config: config}
}

func (app *ExporterApp) Run() {
	kubeConfig, err := app.connectCluster()
	if err != nil {
		logger.Fatalf("Cannot connect cluster: %s", err)
	}
	logger.Info("Cluster connected")

	topology, err := NewModelBuilder(app.config).BuildForKubeConfig(kubeConfig)
	if err != nil {
		logger.Fatalf("Cannot build data model", err)
	}

	fmt := NewFormatterForConfig(app.config)
	output := fmt.Format(topology)
	reporter := NewFileReporter(app.config)
	reporter.Report(output)
}

func (app *ExporterApp) initKubeconfig() *string {
	if home := app.homeDir(); home != "" {
		return flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "")
	} else {
		return flag.String("kubeconfig", "", "")
	}
}

func (app *ExporterApp) connectCluster() (*rest.Config, error) {
	kubeconfig := app.initKubeconfig()
	//Load config for Openshift's go-client from kubeconfig file
	return clientcmd.BuildConfigFromFlags("", *kubeconfig)
}

func (app *ExporterApp) homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
