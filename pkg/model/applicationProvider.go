package model

type ApplicationProvider interface {
	ApplicationConfigs() []ApplicationConfig
}
