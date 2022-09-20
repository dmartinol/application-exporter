package model

import (
	"fmt"
	"strings"

	appsv1T "github.com/openshift/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeploymentConfig struct {
	Delegate appsv1T.DeploymentConfig
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
func (d DeploymentConfig) Icon() string {
	return "images/deployment.png"
}
func (d DeploymentConfig) StatusColor() (string, bool) {
	return "", false
}
func (d DeploymentConfig) OwnerReferences() []metav1.OwnerReference {
	return d.Delegate.OwnerReferences
}
func (d DeploymentConfig) IsOwnerOf(owner metav1.OwnerReference) bool {
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
	for i := 0; i < len(d.Delegate.Spec.Template.Spec.Containers); i++ {
		c := d.Delegate.Spec.Template.Spec.Containers[i]
		apps = append(apps, ApplicationConfig{ApplicationName: c.Name, ImageName: c.Image})
	}
	return apps
}
