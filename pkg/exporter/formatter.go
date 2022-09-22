package exporter

import (
	"fmt"
	"sort"
	"strings"

	logger "github.com/dmartinol/deployment-exporter/pkg/log"
	"github.com/dmartinol/deployment-exporter/pkg/model"
)

type ByNamespaceName []model.NamespaceModel

func (a ByNamespaceName) Len() int           { return len(a) }
func (a ByNamespaceName) Less(i, j int) bool { return a[i].Name() < a[j].Name() }
func (a ByNamespaceName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type Formatter struct {
	contentType ContentType
}

func NewFormatterForContentType(contentType ContentType) Formatter {
	return Formatter{contentType: contentType}
}

func (f Formatter) Format(topologyModel *model.TopologyModel) *strings.Builder {
	logger.Infof("Received formatting request by %s", f.contentType)
	switch f.contentType {
	case Text:
		return f.text(topologyModel)
	case CSV:
		return f.csv(topologyModel)
	}
	var sb = &strings.Builder{}
	sb.WriteString(fmt.Sprintf("Unmanaged content type %s", f.contentType))
	return sb
}

func (f Formatter) sortedNamespaces(topologyModel *model.TopologyModel) []model.NamespaceModel {
	namespaces := topologyModel.AllNamespaces()
	sort.Sort(ByNamespaceName(namespaces))
	return namespaces
}

func appendNewLine(sb *strings.Builder, format string, args ...any) {
	sb.WriteString(fmt.Sprintf(format+"\n", args...))
	// TODO \n
}

func (f Formatter) text(topologyModel *model.TopologyModel) *strings.Builder {
	var sb = &strings.Builder{}

	for _, namespace := range f.sortedNamespaces(topologyModel) {
		for _, applicationProvider := range namespace.AllApplicationProviders() {
			appendNewLine(sb, "===============\nNamespace: %s\nApplication: %s", namespace.Name(), applicationProvider.(model.Resource).Name())
			for _, applicationConfig := range applicationProvider.ApplicationConfigs() {
				applicationImage, ok := topologyModel.ImageByName(applicationConfig.ImageName)
				if ok {
					appendNewLine(sb, "Image name: %s\n", applicationImage.ImageName())
					appendNewLine(sb, "Image version: %s\n", applicationImage.ImageVersion())
					appendNewLine(sb, "Image full name: %s\n", applicationImage.ImageFullName())
				} else {
					appendNewLine(sb, "Image name: %s\n", "NA")
					appendNewLine(sb, "Image version: %s\n", "NA")
					appendNewLine(sb, "Image full name: %s\n", applicationConfig.ImageName)
				}
			}
		}
	}
	return sb
}

func (f Formatter) csv(topologyModel *model.TopologyModel) *strings.Builder {
	var sb = &strings.Builder{}
	appendNewLine(sb, "namespace, application, imageName, imageVersion, fullImageName")

	for _, namespace := range f.sortedNamespaces(topologyModel) {
		for _, applicationProvider := range namespace.AllApplicationProviders() {
			logger.Infof("## %s %s", applicationProvider.(model.Resource).Kind(), applicationProvider.(model.Resource).Name())
			for _, applicationConfig := range applicationProvider.ApplicationConfigs() {
				var record []string
				record = append(record, namespace.Name(), applicationConfig.ApplicationName)
				applicationImage, ok := topologyModel.ImageByName(applicationConfig.ImageName)
				if ok {
					record = append(record, applicationImage.ImageName(), applicationImage.ImageVersion(), applicationImage.ImageFullName())
				} else {
					record = append(record, applicationConfig.ImageName, "NA", applicationConfig.ImageName)
				}
				appendNewLine(sb, "%s", strings.Join(record, ","))
			}
		}
	}
	return sb
}
