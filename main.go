package main

import (
	exp "github.com/dmartinol/deployment-exporter/pkg/exporter"
	log "github.com/dmartinol/deployment-exporter/pkg/log"
)

func main() {
	config := exp.NewConfig()
	log.InitLogger(config.RunInVM(), config.LogLevel())
	log.Infof("Config is %v+", config)

	var exporter exp.Exporter
	if config.RunAsScript() {
		exporter = exp.NewExporterApp(config)
	} else {
		exporter = exp.NewExporterService(config)
	}
	exporter.Run()
}
