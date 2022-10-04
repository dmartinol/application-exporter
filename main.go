package main

import (
	"os"

	exp "github.com/dmartinol/application-exporter/pkg/exporter"
	logger "github.com/dmartinol/application-exporter/pkg/log"
)

var BuildVersion = "development"

func main() {
	config := exp.NewConfig()
	logger.InitLogger(config.RunInVM(), config.LogLevel())
	logger.Infof("The version of %s is : %s\n", os.Args[0], BuildVersion)
	logger.Infof("Config is %v+", config)

	var exporter exp.Exporter
	if config.RunAsScript() {
		exporter = exp.NewExporterApp(config)
	} else {
		exporter = exp.NewExporterService(config)
	}
	exporter.Start()
}
