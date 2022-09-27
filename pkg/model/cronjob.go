package model

import (
	"fmt"
	"strings"

	k8sBatchV1 "k8s.io/api/batch/v1"
	k8sMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CronJob struct {
	Delegate k8sBatchV1.CronJob
}

func (c CronJob) Kind() string {
	return "CronJob"
}
func (c CronJob) Id() string {
	return fmt.Sprintf("cronjob %s", c.Delegate.Name)
}
func (c CronJob) Name() string {
	return c.Delegate.Name
}
func (c CronJob) Label() string {
	return c.Delegate.Name
}

func (c CronJob) OwnerReferences() []k8sMetaV1.OwnerReference {
	return c.Delegate.OwnerReferences
}
func (c CronJob) IsOwnerOf(owner k8sMetaV1.OwnerReference) bool {
	switch owner.Kind {
	case "Job":
		return strings.HasPrefix(owner.Name, c.Name())
	}
	return false
}

func (c CronJob) ApplicationConfigs() []ApplicationConfig {
	var apps []ApplicationConfig
	for _, c := range c.Delegate.Spec.JobTemplate.Spec.Template.Spec.Containers {
		apps = append(apps, ApplicationConfig{ContainerName: c.Name, ImageName: c.Image, Resources: c.Resources})
	}
	return apps
}
