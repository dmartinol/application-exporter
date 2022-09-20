package model

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Deployment struct {
	Delegate v1.Deployment
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

func (d Deployment) Icon() string {
	return "images/deployment.png"
}
func (d Deployment) StatusColor() (string, bool) {
	return "", false
}
func (d Deployment) OwnerReferences() []metav1.OwnerReference {
	return d.Delegate.OwnerReferences
}
func (d Deployment) IsOwnerOf(owner metav1.OwnerReference) bool {
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
	for i := 0; i < len(d.Delegate.Spec.Template.Spec.Containers); i++ {
		c := d.Delegate.Spec.Template.Spec.Containers[i]
		apps = append(apps, ApplicationConfig{ApplicationName: c.Name, ImageName: c.Image})
	}
	return apps
}
