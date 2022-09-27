package model

import (
	k8sMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Resource interface {
	Kind() string
	Id() string
	Name() string
	Label() string

	OwnerReferences() []k8sMetaV1.OwnerReference
	IsOwnerOf(owner k8sMetaV1.OwnerReference) bool
}

type ApplicationProvider interface {
	ApplicationConfigs() []ApplicationConfig
}
