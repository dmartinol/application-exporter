package model

import (
	"strings"

	logger "github.com/dmartinol/application-exporter/pkg/log"
	k8sMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NamespaceModel struct {
	name            string
	resourcesByKind map[string][]Resource
}

func (namespace NamespaceModel) Name() string {
	return namespace.name
}

func (namespace NamespaceModel) LookupByKindAndId(kind string, id string) Resource {
	for _, resource := range namespace.resourcesByKind[kind] {
		if strings.Compare(id, resource.Id()) == 0 {
			return resource
		}
	}

	return nil
}

func (namespace NamespaceModel) AddResource(resource Resource) bool {
	if namespace.LookupByKindAndId(resource.Kind(), resource.Id()) == nil {
		logger.Debugf("Adding resource %s of kind %s", resource.Name(), resource.Kind())
		namespace.resourcesByKind[resource.Kind()] = append(namespace.resourcesByKind[resource.Kind()], resource)
		return true
	}
	logger.Debugf("Skipped existing resource %s of kind %s", resource.Name(), resource.Kind())
	return false
}
func (namespace NamespaceModel) LookupOwner(owner k8sMetaV1.OwnerReference) Resource {
	for _, resources := range namespace.resourcesByKind {
		for _, resource := range resources {
			if resource.IsOwnerOf(owner) {
				return resource
			}
		}
	}
	return nil
}

func (namespace NamespaceModel) ResourcesByKind(kind string) []Resource {
	return namespace.resourcesByKind[kind]
}
func (namespace NamespaceModel) AllApplicationProviders() []ApplicationProvider {
	applicationProviders := make([]ApplicationProvider, 0)
	for _, resource := range namespace.AllResources() {
		if applicationProvider, ok := resource.(ApplicationProvider); ok {
			applicationProviders = append(applicationProviders, applicationProvider)
		}
	}
	return applicationProviders
}
func (namespace NamespaceModel) AllResources() []Resource {
	resources := make([]Resource, 0)
	for kind := range namespace.resourcesByKind {
		resources = append(resources, namespace.resourcesByKind[kind]...)
	}
	return resources
}
func (namespace NamespaceModel) AllPodsOf(parent Resource) []Pod {
	var children []Pod
	for _, pod := range namespace.ResourcesByKind("Pod") {
		for _, ref := range pod.OwnerReferences() {
			if parent.IsOwnerOf(ref) {
				children = append(children, pod.(Pod))
			}
		}
	}
	return children
}
