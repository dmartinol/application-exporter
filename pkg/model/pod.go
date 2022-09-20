package model

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CompletedColor = "#66ff33"
	RunningColor   = "#00ffff"
	FailedColor    = "#ff3300"
)

type Pod struct {
	Delegate v1.Pod
}

func (d Pod) Kind() string {
	return "Pod"
}
func (p Pod) Id() string {
	return fmt.Sprintf("pod %s", p.Delegate.Name)
}
func (p Pod) Name() string {
	return p.Delegate.Name
}
func (p Pod) Label() string {
	return p.Delegate.Name
}
func (p Pod) Icon() string {
	return "images/pod.png"
}
func (p Pod) OwnerReferences() []metav1.OwnerReference {
	return p.Delegate.OwnerReferences
}
func (p Pod) IsOwnerOf(owner metav1.OwnerReference) bool {
	return false
}

func (p Pod) StatusColor() (string, bool) {
	switch p.Delegate.Status.Phase {
	case "Succeeded":
		return CompletedColor, true
	case "Running":
		if p.isReady() {
			return RunningColor, true
		}
	}
	return FailedColor, true
}

func (p Pod) isReady() bool {
	for _, c := range p.Delegate.Status.Conditions {
		if strings.Compare(string(c.Type), "Ready") == 0 {
			return strings.Compare(string(c.Status), "True") == 0
		}
	}
	return false
}
