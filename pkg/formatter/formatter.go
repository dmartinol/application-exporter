package formatter

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dmartinol/application-exporter/pkg/config"
	logger "github.com/dmartinol/application-exporter/pkg/log"
	"github.com/dmartinol/application-exporter/pkg/model"
	k8sCoreV1 "k8s.io/api/core/v1"
)

type ByNamespaceName []model.NamespaceModel

func (a ByNamespaceName) Len() int           { return len(a) }
func (a ByNamespaceName) Less(i, j int) bool { return a[i].Name() < a[j].Name() }
func (a ByNamespaceName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type Formatter struct {
	config *config.Config
}

func NewFormatterForConfig(config *config.Config) Formatter {
	return Formatter{config: config}
}

func (f Formatter) Format(topologyModel *model.TopologyModel) *strings.Builder {
	logger.Infof("Received formatting request by %s", f.config.ContentType())
	switch f.config.ContentType() {
	case config.Text:
		return f.text(topologyModel)
	case config.CSV:
		return f.csv(topologyModel)
	}
	var sb = &strings.Builder{}
	sb.WriteString(fmt.Sprintf("Unmanaged content type %s", f.config.ContentType()))
	return sb
}

func SortedNamespaces(topologyModel *model.TopologyModel) []model.NamespaceModel {
	namespaces := topologyModel.AllNamespaces()
	sort.Sort(ByNamespaceName(namespaces))
	return namespaces
}

func appendNewLine(sb *strings.Builder, format string, args ...any) {
	sb.WriteString(fmt.Sprintf(format+"\n", args...))
}

func CpuLimits(resources k8sCoreV1.ResourceRequirements) string {
	if val, ok := resources.Limits[k8sCoreV1.ResourceCPU]; ok {
		return val.String()
	}
	return "NA"
}
func MemoryLimits(resources k8sCoreV1.ResourceRequirements) string {
	if val, ok := resources.Limits[k8sCoreV1.ResourceMemory]; ok {
		return val.String()
	}
	return "NA"
}
func CpuRequests(resources k8sCoreV1.ResourceRequirements) string {
	if val, ok := resources.Requests[k8sCoreV1.ResourceCPU]; ok {
		return val.String()
	}
	return "NA"
}
func MemoryRequests(resources k8sCoreV1.ResourceRequirements) string {
	if val, ok := resources.Requests[k8sCoreV1.ResourceMemory]; ok {
		return val.String()
	}
	return "NA"
}

func CpuUsage(usage k8sCoreV1.ResourceList) string {
	return usage.Cpu().String()
}
func MemoryUsage(usage k8sCoreV1.ResourceList) string {
	return usage.Memory().String()
}

func (f Formatter) text(topologyModel *model.TopologyModel) *strings.Builder {
	var sb = &strings.Builder{}

	for _, namespace := range SortedNamespaces(topologyModel) {
		for _, applicationProvider := range namespace.AllApplicationProviders() {
			appendNewLine(sb, "===============\nNamespace: %s\nApplication: %s (%s)", namespace.Name(), applicationProvider.(model.Resource).Name(), applicationProvider.(model.Resource).Kind())
			for _, applicationConfig := range applicationProvider.ApplicationConfigs() {
				appendNewLine(sb, "Container name: %s\n", applicationConfig.ContainerName)
				applicationImage, ok := topologyModel.ImageByName(applicationConfig.ImageName)
				if ok {
					appendNewLine(sb, "Image name: %s", applicationImage.ImageName())
					appendNewLine(sb, "Image version: %s", applicationImage.ImageVersion())
					appendNewLine(sb, "Image full name: %s", applicationImage.ImageFullName())
				} else {
					appendNewLine(sb, "Image name: %s", "NA")
					appendNewLine(sb, "Image version: %s", "NA")
					appendNewLine(sb, "Image full name: %s", applicationConfig.ImageName)
				}
				if f.config.WithResources() {
					res := applicationConfig.Resources
					appendNewLine(sb, "Limits: %s CPU, %s memory\nRequests: %s CPU, %s memory", CpuLimits(res), MemoryLimits(res), CpuRequests(res), MemoryRequests(res))

					for _, pod := range namespace.AllPodsOf(applicationProvider.(model.Resource)) {
						if pod.IsRunning() {
							appendNewLine(sb, "\nPod name: %s", pod.Name())
							usage := pod.UsageForContainer(applicationConfig.ContainerName)
							if usage != nil {
								appendNewLine(sb, "Usage: %s CPU, %s memory", CpuUsage(usage), MemoryUsage(usage))
							} else {
								usage = pod.UsageForContainer(pod.Name())
								if usage != nil {
									appendNewLine(sb, "Usage: %s CPU, %s memory", CpuUsage(usage), MemoryUsage(usage))
								} else {
									appendNewLine(sb, "No Usage metrics")
								}
							}
						}
					}
				}
			}
		}
	}
	return sb
}

func (f Formatter) csv(topologyModel *model.TopologyModel) *strings.Builder {
	var sb = &strings.Builder{}
	if f.config.WithResources() {
		appendNewLine(sb, "namespace, application, container, imageName, imageVersion, fullImageName, CPU limits, memory limits, CPU requests, memory requests, pod, CPU usage, memory usage")
	} else {
		appendNewLine(sb, "namespace, application, container, imageName, imageVersion, fullImageName")
	}

	for _, namespace := range SortedNamespaces(topologyModel) {
		for _, applicationProvider := range namespace.AllApplicationProviders() {
			logger.Debugf("## %s %s", applicationProvider.(model.Resource).Kind(), applicationProvider.(model.Resource).Name())
			for _, applicationConfig := range applicationProvider.ApplicationConfigs() {
				var record []string
				record = append(record, namespace.Name(), applicationProvider.(model.Resource).Name(), applicationConfig.ContainerName)
				applicationImage, ok := topologyModel.ImageByName(applicationConfig.ImageName)
				if ok {
					record = append(record, applicationImage.ImageName(), applicationImage.ImageVersion(), applicationImage.ImageFullName())
				} else {
					record = append(record, applicationConfig.ImageName, "NA", applicationConfig.ImageName)
				}
				if f.config.WithResources() {
					res := applicationConfig.Resources
					record = append(record, CpuLimits(res), MemoryLimits(res), CpuRequests(res), MemoryRequests(res))

					headerRecord := make([]string, len(record))
					copy(headerRecord, record)

					for _, pod := range namespace.AllPodsOf(applicationProvider.(model.Resource)) {
						if pod.IsRunning() {
							usage := pod.UsageForContainer(applicationConfig.ContainerName)
							if usage != nil {
								record = append(headerRecord, pod.Name(), CpuUsage(usage), MemoryUsage(usage))
							} else {
								usage = pod.UsageForContainer(pod.Name())
								if usage != nil {
									record = append(headerRecord, pod.Name(), CpuUsage(usage), MemoryUsage(usage))
								} else {
									record = append(headerRecord, pod.Name(), "NA", "NA")
								}
							}
							appendNewLine(sb, "%s", strings.Join(record, ","))
						}
					}
				} else {
					appendNewLine(sb, "%s", strings.Join(record, ","))
				}
			}
		}
	}
	return sb
}
