package formatter

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	logger "github.com/dmartinol/deployment-exporter/pkg/log"
	"github.com/dmartinol/deployment-exporter/pkg/model"
)

type ContentType int64

const (
	Text ContentType = iota
	CSV
	JSON
	YAML
)

func (t ContentType) String() string {
	switch t {
	case Text:
		return "text"
	case CSV:
		return "CSV"
	case JSON:
		return "JSON"
	case YAML:
		return "YAML"
	}
	return "unknown"
}
func ContentTypeFromString(contentType string) ContentType {
	switch contentType {
	case "text":
		return Text
	case "CSV":
		return CSV
	case "JSON":
		return JSON
	case "YAML":
		return YAML
	}
	return Text
}

type Formatter struct {
	TopologyModel model.TopologyModel
}

type ByNamespaceName []model.NamespaceModel

func (a ByNamespaceName) Len() int           { return len(a) }
func (a ByNamespaceName) Less(i, j int) bool { return a[i].Name() < a[j].Name() }
func (a ByNamespaceName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (f Formatter) Format(contentType ContentType, rw http.ResponseWriter) {
	logger.Infof("Received formatting request by %s", contentType)
	switch contentType {
	case Text:
		f.text(rw)
	case CSV:
		f.CSV(rw)
	case JSON:
	case YAML:
		logger.Warnf("Unmanaged type %s", contentType)
	}
}

func (f Formatter) sortedNamespaces() []model.NamespaceModel {
	namespaces := f.TopologyModel.AllNamespaces()
	sort.Sort(ByNamespaceName(namespaces))
	return namespaces
}

func (f Formatter) text(rw http.ResponseWriter) {
	var sb = &strings.Builder{}

	for _, namespace := range f.sortedNamespaces() {
		for _, applicationProvider := range namespace.AllApplicationProviders() {
			sb.WriteString(fmt.Sprintf("===============\nNamespace: %s\nApplication: %s\n", namespace.Name(), applicationProvider.(model.Resource).Name()))
			for _, applicationConfig := range applicationProvider.ApplicationConfigs() {
				applicationImage, ok := f.TopologyModel.ImageByName(applicationConfig.ImageName)
				if ok {
					sb.WriteString(fmt.Sprintf("Image name: %s\n", applicationImage.ImageName()))
					sb.WriteString(fmt.Sprintf("Image version: %s\n", applicationImage.ImageVersion()))
					sb.WriteString(fmt.Sprintf("Image full name: %s\n", applicationImage.ImageFullName()))
				} else {
					sb.WriteString(fmt.Sprintf("Image name: %s\n", "NA"))
					sb.WriteString(fmt.Sprintf("Image version: %s\n", "NA"))
					sb.WriteString(fmt.Sprintf("Image full name: %s\n", applicationConfig.ImageName))
				}
			}
		}
	}

	rw.WriteHeader(http.StatusOK)
	rw.Header().Set("Content-Type", "application/text")
	rw.Write([]byte(sb.String()))
}

func (f Formatter) CSV(rw http.ResponseWriter) {
	// file, err := os.Create("exporter.csv")
	// if err != nil {
	// 	log.Fatalln("Error creating CSV file:", err)
	// }
	// defer file.Close()

	// w := csv.NewWriter(file)
	// defer w.Flush()

	var items [][]string
	for _, namespace := range f.sortedNamespaces() {
		for _, applicationProvider := range namespace.AllApplicationProviders() {
			logger.Infof("## %s %s", applicationProvider.(model.Resource).Kind(), applicationProvider.(model.Resource).Name())
			for _, applicationConfig := range applicationProvider.ApplicationConfigs() {
				var record []string
				record = append(record, namespace.Name(), applicationConfig.ApplicationName)
				applicationImage, ok := f.TopologyModel.ImageByName(applicationConfig.ImageName)
				if ok {
					record = append(record, applicationImage.ImageName(), applicationImage.ImageVersion(), applicationImage.ImageFullName())
				} else {
					record = append(record, applicationConfig.ImageName, "NA", applicationConfig.ImageName)
				}
				items = append(items, record)
			}
		}
	}

	rw.Header().Set("Content-Type", "text/csv")
	rw.Header().Set("Content-Disposition", "attachment;filename=inventory.csv")
	wr := csv.NewWriter(rw)

	header := []string{"namespace", "application", "imageName", "imageVersion", "fullImageName"}
	if err := wr.Write(header); err != nil {
		log.Fatalln("Error writing header to csv:", err)
	}
	if err := wr.WriteAll(items); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}
