package model

import (
	"fmt"
	"strings"

	appsV1 "github.com/openshift/api/apps/v1"
	k8sMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeploymentConfig struct {
	Delegate appsV1.DeploymentConfig
}

func (d DeploymentConfig) Kind() string {
	return "DeploymentConfig"
}
func (d DeploymentConfig) Id() string {
	return fmt.Sprintf("deploymentconfig %s", d.Delegate.Name)
}
func (d DeploymentConfig) Name() string {
	return d.Delegate.Name
}
func (d DeploymentConfig) Label() string {
	return d.Delegate.Name
}

func (d DeploymentConfig) OwnerReferences() []k8sMetaV1.OwnerReference {
	return d.Delegate.OwnerReferences
}
func (d DeploymentConfig) IsOwnerOf(owner k8sMetaV1.OwnerReference) bool {
	switch owner.Kind {
	case "DeploymentConfig":
		return strings.Compare(owner.Name, d.Name()) == 0
	case "ReplicationController":
		return strings.HasPrefix(owner.Name, d.Name())
	}
	return false
}

func (d DeploymentConfig) ApplicationConfigs() []ApplicationConfig {
	var apps []ApplicationConfig
	for _, c := range d.Delegate.Spec.Template.Spec.Containers {
		apps = append(apps, ApplicationConfig{ContainerName: c.Name, ImageName: c.Image, Resources: c.Resources})
	}
	return apps
}
