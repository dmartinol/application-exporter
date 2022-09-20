package model

import "strings"

type ApplicationConfig struct {
	ApplicationName string
	ImageName       string
}

func (a ApplicationConfig) IsImageStream() bool {
	return strings.Contains(a.ImageName, "@sha")
}
func (a ApplicationConfig) ImageStreamId() string {
	return a.ImageName[strings.LastIndex(a.ImageName, "/")+1:]
}
