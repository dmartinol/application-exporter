package main

import (
	"os"

	"github.com/dmartinol/application-exporter/pkg/config"
	exp "github.com/dmartinol/application-exporter/pkg/exporter"
	logger "github.com/dmartinol/application-exporter/pkg/log"
	"github.com/dmartinol/application-exporter/pkg/monitor"
)

var BuildVersion = "development"

func main() {
	config := config.NewConfig()
	logger.InitLogger(config.RunInVM(), config.LogLevel())
	logger.Infof("The version of %s is : %s\n", os.Args[0], BuildVersion)
	logger.Infof("Config is %v+", config)

	var exporter exp.Exporter
	if config.RunAsScript() {
		exporter = exp.NewExporterApp(config)
	} else if config.RunAsMonitoring() {
		exporter = monitor.NewExporterMetrics(config)
	} else if config.RunAsService() {
		exporter = exp.NewExporterService(config)
	}
	exporter.Start()
}
