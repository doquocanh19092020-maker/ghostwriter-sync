package main

import (
	"database/sql"
	"errors"

	havoc "github.com/InfinityCurveLabs/havoc-sdk/backend"
)

type GhostWriterPlugin struct {
	havoc.Plugin

	ghostwriter struct {
		executionInput map[string]string
		client         *GhostClient
		database       *sql.DB
	}
}

// export the plugin object
var Plugin GhostWriterPlugin

var (
	PluginName    = "kaine.ghostwriter"
	PluginVersion string
	PluginBuildId string
)

// Register
// is the first function to be called once the
// plugin has been loaded into the havoc teamserver
func (hc *GhostWriterPlugin) Register(server any) (map[string]any, error) {
	// cast the given havoc interface to
	// the internal plugin HavocInterface
	hc.IHavocCore = server.(havoc.IHavocCore)

	// return the plugin metadata
	return map[string]any{
		"name":        PluginName,
		"type":        havoc.PluginTypeManagement,
		"description": "Kaine agent ghostwriter plugin",
		"version":     PluginVersion,
		"author":      "C5pider",
	}, nil
}

func (hc *GhostWriterPlugin) UnRegister() error {
	_ = hc.TraceAgentEventRemove(hc.KaineTraceEventCallback)
	return nil
}

func (hc *GhostWriterPlugin) ManagerRegister() error {
	_ = hc.TraceAgentEventRegister(hc.KaineTraceEventCallback)

	hc.Profile()

	return nil
}

func (hc *GhostWriterPlugin) KaineTraceEventCallback(event, uuid string, data map[string]any) {
	var (
		agent string
		err   error
	)

	if agent, err = hc.AgentType(uuid); err != nil {
		return
	}

	// check if it is our target agent type we want to log
	if agent != "Kaine" {
		return
	}

	if hc.ghostwriter.client != nil {
		hc.GhostWriter(event, uuid, data)
	}
}

func (hc *GhostWriterPlugin) Call(user string, arguments any) (any, error) {
	return nil, errors.New("unknown command")
}
