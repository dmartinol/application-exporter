package monitor

import (
	"fmt"
	"net/http"

	"github.com/dmartinol/application-exporter/pkg/config"
	"github.com/dmartinol/application-exporter/pkg/exporter"
	"github.com/dmartinol/application-exporter/pkg/formatter"
	logger "github.com/dmartinol/application-exporter/pkg/log"
	"github.com/dmartinol/application-exporter/pkg/model"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

/*
* References:
* https://github.com/openshift/cluster-version-operator
* https://prometheus.io/docs/guides/go-application/
 */

type ExporterMetrics struct {
	config             *config.Config
	appVersion         *prometheus.GaugeVec
	appResourcesConfig *prometheus.GaugeVec
	appResourcesUsage  *prometheus.GaugeVec
}

var router = mux.NewRouter()

func NewExporterMetrics(config *config.Config) *ExporterMetrics {
	exporterMetrics := ExporterMetrics{}
	// TBD maybe not used
	exporterMetrics.config = config
	exporterMetrics.appVersion = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "application_version",
		Help: `.`,
	}, []string{"environment", "namespace", "application", "type", "container", "image", "version", "full_image"})
	exporterMetrics.appResourcesConfig = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "application_resources_config",
		Help: `.`,
	}, []string{"environment", "namespace", "application", "type", "container", "cpu_limits", "memory_limits", "cpu_requests", "memory_requests"})
	exporterMetrics.appResourcesUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "application_resources_usage",
		Help: `.`,
	}, []string{"environment", "namespace", "application", "type", "pod", "container", "cpu_usage", "memory_usage"})

	prometheus.Register(&exporterMetrics)

	return &exporterMetrics
}

func (s *ExporterMetrics) Start() {
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

// Collect implements prometheus.Collector
func (em *ExporterMetrics) Collect(ch chan<- prometheus.Metric) {
	logger.Infof("Collect invoked")

	// Creates a REST service exporter but does not start it
	exporterService := exporter.NewExporterService(em.config)
	runner := exporterService.NewRunner(em.config, nil, nil)

	kubeConfig, err := runner.Connect()
	if err != nil {
		logger.Fatalf("Cannot connect cluster: %s", err)
		// TBD
		// return err
	}

	logger.Info("Cluster connected")
	topology, err := runner.Collect(kubeConfig)
	if err != nil {
		logger.Fatalf("Cannot collect metrics from cluster: %s", err)
		// TBD
		// return err
	}

	for _, namespace := range formatter.SortedNamespaces(topology) {
		for _, applicationProvider := range namespace.AllApplicationProviders() {
			logger.Debugf("## %s %s", applicationProvider.(model.Resource).Kind(), applicationProvider.(model.Resource).Name())
			for _, applicationConfig := range applicationProvider.ApplicationConfigs() {
				g := em.applicationVersionMetric(topology, namespace.Name(), applicationProvider.(model.Resource), applicationConfig)
				logger.Debugf("Adding to ch: %s", g.Desc())
				ch <- g

				if em.config.WithResources() {
					g = em.resourcesConfigMetric(namespace.Name(), applicationProvider.(model.Resource), applicationConfig)
					logger.Debugf("Adding to ch: %s", g.Desc())
					ch <- g

					for _, g := range em.resourcesUsageMetric(namespace, applicationProvider.(model.Resource), applicationConfig) {
						logger.Debugf("Adding to ch: %s", g.Desc())
						ch <- g
					}
				}
			}
		}
	}
}

// Describe implements prometheus.Collector
func (m *ExporterMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.appVersion.WithLabelValues("", "", "", "", "", "", "", "").Desc()
}

func (em *ExporterMetrics) applicationVersionMetric(topology *model.TopologyModel, namespace string, application model.Resource, applicationConfig model.ApplicationConfig) prometheus.Gauge {
	var record []string
	record = append(record, em.config.Environment(), namespace, application.Name(), application.Kind(), applicationConfig.ContainerName)
	applicationImage, ok := topology.ImageByName(applicationConfig.ImageName)
	if ok {
		record = append(record, applicationImage.ImageName(), applicationImage.ImageVersion(), applicationImage.ImageFullName())
	} else {
		record = append(record, applicationConfig.ImageName, "NA", applicationConfig.ImageName)
	}

	g := em.appVersion.WithLabelValues(record...)
	// TBD
	g.Set(0)
	return g
}

func (em *ExporterMetrics) resourcesConfigMetric(namespace string, application model.Resource, applicationConfig model.ApplicationConfig) prometheus.Gauge {
	var record []string
	res := applicationConfig.Resources
	record = append(record, em.config.Environment(), namespace, application.Name(), application.Kind(), applicationConfig.ContainerName)
	record = append(record, formatter.CpuLimits(res), formatter.MemoryLimits(res), formatter.CpuRequests(res), formatter.MemoryRequests(res))
	g := em.appResourcesConfig.WithLabelValues(record...)
	// TBD
	g.Set(0)
	return g
}

func (em *ExporterMetrics) resourcesUsageMetric(namespace model.NamespaceModel, application model.Resource, applicationConfig model.ApplicationConfig) []prometheus.Gauge {
	var metrics []prometheus.Gauge

	for _, pod := range namespace.AllPodsOf(application) {
		if pod.IsRunning() {
			var record []string
			record = append(record, em.config.Environment(), namespace.Name(), application.Name(), application.Kind(), pod.Name(), applicationConfig.ContainerName)
			usage := pod.UsageForContainer(applicationConfig.ContainerName)
			if usage != nil {
				record = append(record, formatter.CpuUsage(usage), formatter.MemoryUsage(usage))
			} else {
				usage = pod.UsageForContainer(pod.Name())
				if usage != nil {
					record = append(record, formatter.CpuUsage(usage), formatter.MemoryUsage(usage))
				} else {
					record = append(record, "NA", "NA")
				}
			}
			g := em.appResourcesUsage.WithLabelValues(record...)
			// TBD
			g.Set(0)

			metrics = append(metrics, g)
		}
	}

	return metrics
}
