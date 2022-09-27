package model

import (
	"strings"

	k8sCoreV1 "k8s.io/api/core/v1"
)

type ApplicationConfig struct {
	ContainerName  string
	ImageName      string
	Resources      k8sCoreV1.ResourceRequirements
	ResourcesUsage k8sCoreV1.ResourceList
}

func (a ApplicationConfig) IsImageStream() bool {
	return strings.Contains(a.ImageName, "@sha")
}
func (a ApplicationConfig) ImageStreamId() string {
	return a.ImageName[strings.LastIndex(a.ImageName, "/")+1:]
}
