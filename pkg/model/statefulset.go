package model

import (
	"fmt"
	"strings"

	k8sCoreV1 "k8s.io/api/apps/v1"
	k8sMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StatefulSet struct {
	Delegate k8sCoreV1.StatefulSet
}

func (s StatefulSet) Kind() string {
	return "StatefulSet"
}
func (s StatefulSet) Id() string {
	return fmt.Sprintf("sts %s", s.Delegate.Name)
}
func (s StatefulSet) Name() string {
	return s.Delegate.Name
}
func (s StatefulSet) Label() string {
	return s.Delegate.Name
}
func (s StatefulSet) Icon() string {
	return "images/sts.png"
}
func (s StatefulSet) StatusColor() (string, bool) {
	return "", false
}
func (s StatefulSet) OwnerReferences() []k8sMetaV1.OwnerReference {
	return s.Delegate.OwnerReferences
}
func (s StatefulSet) IsOwnerOf(owner k8sMetaV1.OwnerReference) bool {
	return strings.Compare(owner.Kind, s.Kind()) == 0 && strings.Compare(owner.Name, s.Name()) == 0
}

func (s StatefulSet) ApplicationConfigs() []ApplicationConfig {
	var apps []ApplicationConfig
	for _, c := range s.Delegate.Spec.Template.Spec.Containers {
		apps = append(apps, ApplicationConfig{ContainerName: c.Name, ImageName: c.Image, Resources: c.Resources})
	}
	return apps
}
