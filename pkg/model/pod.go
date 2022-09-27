package model

import (
	"fmt"
	"strings"

	logger "github.com/dmartinol/application-exporter/pkg/log"
	k8sCoreV1 "k8s.io/api/core/v1"
	k8sMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sMetricsV1Beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

const (
	CompletedColor = "#66ff33"
	RunningColor   = "#00ffff"
	FailedColor    = "#ff3300"
)

type Pod struct {
	Delegate   k8sCoreV1.Pod
	PodMetrics *k8sMetricsV1Beta1.PodMetrics
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
func (p Pod) OwnerReferences() []k8sMetaV1.OwnerReference {
	return p.Delegate.OwnerReferences
}
func (p Pod) IsOwnerOf(owner k8sMetaV1.OwnerReference) bool {
	return false
}

func (p Pod) StatusColor() (string, bool) {
	switch p.Delegate.Status.Phase {
	case k8sCoreV1.PodSucceeded:
		return CompletedColor, true
	case k8sCoreV1.PodRunning:
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
	return p.Delegate.Status.Phase == k8sCoreV1.PodRunning
}

func (p *Pod) SetMetrics(podMetrics *k8sMetricsV1Beta1.PodMetrics) {
	p.PodMetrics = podMetrics
}

func (p Pod) UsageForContainer(containerName string) k8sCoreV1.ResourceList {
	if p.IsRunning() {
		podMetrics := p.PodMetrics
		if podMetrics != nil {
			for _, c := range p.PodMetrics.Containers {
				if c.Name == containerName {
					return c.Usage
				} else if c.Name != "POD" {
					logger.Infof("No match %s, %s, %s", p.Name(), containerName, c.Name)
				}
			}
		}
	}
	return nil
}
