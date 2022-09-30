package model

import "sync"

type TopologyModel struct {
	namespacesByName map[string]*NamespaceModel
	imageByName      map[string]ApplicationImage
}

func NewTopologyModel() *TopologyModel {
	var topology TopologyModel
	topology.namespacesByName = make(map[string]*NamespaceModel)
	topology.imageByName = make(map[string]ApplicationImage)
	return &topology
}

var mutex = sync.RWMutex{}

func (topology TopologyModel) AddNamespace(name string) *NamespaceModel {
	mutex.Lock()
	namespace := NamespaceModel{name: name, resourcesByKind: make(map[string][]Resource)}
	topology.namespacesByName[name] = &namespace
	mutex.Unlock()
	return &namespace
}
func (topology TopologyModel) NamespaceByName(name string) *NamespaceModel {
	return topology.namespacesByName[name]
}
func (topology TopologyModel) AllNamespaces() []NamespaceModel {
	namespaces := make([]NamespaceModel, 0, len(topology.namespacesByName))
	for _, namespace := range topology.namespacesByName {
		namespaces = append(namespaces, *namespace)
	}
	return namespaces
}
func (topology TopologyModel) AddImage(imageName string, image ApplicationImage) {
	topology.imageByName[imageName] = image
}
func (topology TopologyModel) ImageByName(imageName string) (ApplicationImage, bool) {
	image, ok := topology.imageByName[imageName]
	return image, ok
}
