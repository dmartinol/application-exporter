package model

import (
	"fmt"
	"strings"

	k8sAppsV1 "k8s.io/api/apps/v1"
	k8sMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DaemonSet struct {
	Delegate k8sAppsV1.DaemonSet
}

func (d DaemonSet) Kind() string {
	return "DaemonSet"
}
func (d DaemonSet) Id() string {
	return fmt.Sprintf("daemonset %s", d.Delegate.Name)
}
func (d DaemonSet) Name() string {
	return d.Delegate.Name
}
func (d DaemonSet) Label() string {
	return d.Delegate.Name
}

func (d DaemonSet) OwnerReferences() []k8sMetaV1.OwnerReference {
	return d.Delegate.OwnerReferences
}
func (d DaemonSet) IsOwnerOf(owner k8sMetaV1.OwnerReference) bool {
	switch owner.Kind {
	case "DaemonSet":
		return strings.Compare(owner.Name, d.Name()) == 0
	}
	return false
}

func (d DaemonSet) ApplicationConfigs() []ApplicationConfig {
	var apps []ApplicationConfig
	for _, c := range d.Delegate.Spec.Template.Spec.Containers {
		apps = append(apps, ApplicationConfig{ContainerName: c.Name, ImageName: c.Image, Resources: c.Resources})
	}
	return apps
}
