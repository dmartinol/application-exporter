package model

import (
	"fmt"
	"strings"

	k8sAppsV1 "k8s.io/api/apps/v1"
	k8sMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Deployment struct {
	Delegate k8sAppsV1.Deployment
}

func (d Deployment) Kind() string {
	return "Deployment"
}
func (d Deployment) Id() string {
	return fmt.Sprintf("deployment %s", d.Delegate.Name)
}
func (d Deployment) Name() string {
	return d.Delegate.Name
}
func (d Deployment) Label() string {
	return d.Delegate.Name
}

func (d Deployment) OwnerReferences() []k8sMetaV1.OwnerReference {
	return d.Delegate.OwnerReferences
}
func (d Deployment) IsOwnerOf(owner k8sMetaV1.OwnerReference) bool {
	switch owner.Kind {
	case "Deployment":
		return strings.Compare(owner.Name, d.Name()) == 0
	case "ReplicaSet":
		return strings.HasPrefix(owner.Name, d.Name())
	}
	return false
}

func (d Deployment) ApplicationConfigs() []ApplicationConfig {
	var apps []ApplicationConfig
	for _, c := range d.Delegate.Spec.Template.Spec.Containers {
		apps = append(apps, ApplicationConfig{ContainerName: c.Name, ImageName: c.Image, Resources: c.Resources})
	}
	return apps
}
