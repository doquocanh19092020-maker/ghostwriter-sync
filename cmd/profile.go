package main

import (
	"database/sql"
)

func (hc *GhostWriterPlugin) Profile() {
	var (
		kaine  map[string]any
		config map[string]any
		ok     bool
		err    error
		domain string
		token  string
		id     int64
	)

	config = hc.Config()

	if _, ok = config["kaine"]; !ok {
		return
	}

	kaine, err = MapKey[map[string]any](hc.Config(), "kaine")
	if err != nil {
		hc.LogDbgError("failed to parse kaine config entry: %v", err)
		return
	}

	if _, ok = kaine["logger"]; !ok {
		return
	}

	config, err = MapKey[map[string]any](kaine, "ghostwriter")
	if err != nil {
		hc.LogFatal("failed to parse kaine logger config entry: %v", err)
		return
	}

	//
	// parse the ghostwriter config now
	//

	if MapExists(config, "domain") {
		domain, err = MapKey[string](config, "domain")
		if err != nil {
			hc.LogFatal("failed to parse kaine logger ghostwriter config domain entry: %v", err)
			return
		}
	} else {
		hc.LogFatal("failed to parse kaine logger ghostwriter config domain entry: not found")
	}

	if MapExists(config, "api-token") {
		token, err = MapKey[string](config, "api-token")
		if err != nil {
			hc.LogFatal("failed to parse kaine logger ghostwriter config token entry: %v", err)
			return
		}
	} else {
		hc.LogFatal("failed to parse kaine logger ghostwriter config token entry: not found")
	}

	if MapExists(config, "operation-id") {
		id, err = MapKey[int64](config, "operation-id")
		if err != nil {
			hc.LogFatal("failed to parse kaine logger ghostwriter config operation-id entry: %v", err)
			return
		}
	} else {
		hc.LogFatal("failed to parse kaine logger ghostwriter config operation-id entry: not found")
	}

	hc.ghostwriter.executionInput = make(map[string]string)

	hc.ghostwriter.client, err = NewGhostClient(domain, token, id)
	if err != nil {
		hc.LogFatal("failed to create ghostwriter client: %v", err)
	}

	hc.ghostwriter.database, err = sql.Open("sqlite3", hc.ConfigPath()+"/kaine/database.db")
	if err != nil {
		hc.LogFatal("failed to open kaine database: %v", err)
	}
}
