package model

import (
	"strings"

	v1 "k8s.io/api/core/v1"
)

type ApplicationConfig struct {
	ContainerName  string
	ImageName      string
	Resources      v1.ResourceRequirements
	ResourcesUsage v1.ResourceList
}

func (a ApplicationConfig) IsImageStream() bool {
	return strings.Contains(a.ImageName, "@sha")
}
func (a ApplicationConfig) ImageStreamId() string {
	return a.ImageName[strings.LastIndex(a.ImageName, "/")+1:]
}
