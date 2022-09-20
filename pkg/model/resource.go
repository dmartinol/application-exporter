package model

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Resource interface {
	Kind() string
	Id() string
	Name() string
	Label() string
	Icon() string
	StatusColor() (string, bool)

	OwnerReferences() []metav1.OwnerReference
	IsOwnerOf(owner metav1.OwnerReference) bool
}
