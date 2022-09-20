package model

import (
	"encoding/json"
	"strings"

	logger "github.com/dmartinol/deployment-exporter/pkg/log"
	"github.com/openshift/api/image/docker10"
	images "github.com/openshift/api/image/v1"
)

type ApplicationImage interface {
	ImageFullName() string
	ImageName() string
	ImageVersion() string
}

type Image struct {
	FullName string
}

func NewImageByRegistry(imageName string) *Image {
	image := &Image{FullName: imageName}
	return image
}
func (i *Image) ImageFullName() string {
	return i.FullName
}
func (i *Image) ImageName() string {
	imageName := i.FullName[strings.LastIndex(i.FullName, "/")+1:]
	if strings.Contains(imageName, ":") {
		imageName = imageName[0:strings.LastIndex(imageName, ":")]
	}
	return imageName
}
func (i *Image) ImageVersion() string {
	if strings.Contains(i.FullName, ":") {
		return strings.Split(i.FullName, ":")[1]
	}
	return "NA"
}

type ImageByStream struct {
	FullName string
	Delegate images.Image
}

func NewImageByStream(imageFullName string, delegate images.Image) *ImageByStream {
	imageByStream := &ImageByStream{Delegate: delegate}
	imageByStream.FullName = imageFullName
	return imageByStream
}
func (i *ImageByStream) ImageFullName() string {
	return i.FullName
}
func (i *ImageByStream) ImageName() string {
	imagePath := i.onlyImagePath()
	imageName := imagePath[strings.LastIndex(imagePath, "/")+1:]
	if strings.Contains(imageName, ":") {
		imageName = imageName[0:strings.LastIndex(imageName, ":")]
	}
	logger.Debugf("Image name of %s is %s", i.imageReference(), imageName)
	return imageName
}
func (i *ImageByStream) ImageVersion() string {
	imageName := i.onlyImagePath()
	imageVersion := "NA"
	if strings.LastIndex(imageName, ":") > strings.LastIndex(imageName, "/") {
		imageVersion = imageName[strings.LastIndex(imageName, ":")+1:]
	} else {
		obj := &docker10.DockerImage{}
		if len(i.Delegate.DockerImageMetadata.Raw) != 0 {
			if err := json.Unmarshal(i.Delegate.DockerImageMetadata.Raw, obj); err != nil {
				logger.Warnf("Cannot unmarshal DockerImageMetadata: %s", err)
			} else {
				logger.Debugf("Config.Labels for %s: %s", imageName, obj.Config.Labels)
				imageVersion = obj.Config.Labels["version"]
			}
		} else {
			logger.Warnf("No DockerImageMetadata for %s", i.Delegate.Name)
		}
	}
	logger.Debugf("Image version of %s is %s", i.imageReference(), imageVersion)
	return imageVersion
}
func (i *ImageByStream) onlyImagePath() string {
	// Remove @sha
	return strings.Split(i.imageReference(), "@")[0]
}
func (i *ImageByStream) imageReference() string {
	return i.Delegate.DockerImageReference
}
