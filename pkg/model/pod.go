package model

import (
	"fmt"
	"strings"

	logger "github.com/dmartinol/application-exporter/pkg/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

const (
	CompletedColor = "#66ff33"
	RunningColor   = "#00ffff"
	FailedColor    = "#ff3300"
)

type Pod struct {
	Delegate   v1.Pod
	PodMetrics *v1beta1.PodMetrics
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
	case v1.PodSucceeded:
		return CompletedColor, true
	case v1.PodRunning:
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
func (p Pod) IsRunning() bool {
	return p.Delegate.Status.Phase == v1.PodRunning
}

func (p *Pod) SetMetrics(podMetrics *v1beta1.PodMetrics) {
	p.PodMetrics = podMetrics
}

func (p Pod) UsageForContainer(containerName string) v1.ResourceList {
	if p.IsRunning() {
		podMetrics := p.PodMetrics
		if podMetrics != nil {
			for _, c := range p.PodMetrics.Containers {
				if c.Name == containerName {
					return c.Usage
				} else {
					logger.Infof("No match %s, %s, %s", p.Name(), containerName, c.Name)
				}
			}
		}
	}
	return nil
}
