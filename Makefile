
VERSION  ?= 0.1
BUILD_ID ?= dev

all:
	go build -buildmode=plugin -trimpath -ldflags="-s -w -X main.PluginVersion=$(VERSION) -X main.PluginBuildId=$(BUILD_ID)" -o=plugin.hc cmd/*
