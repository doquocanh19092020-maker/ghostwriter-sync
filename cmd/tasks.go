package main

import (
	"database/sql"
	"errors"
)

func (hc *GhostWriterPlugin) TaskExists(uuid string, task int) (bool, error) {
	var (
		err    error
		exists bool
	)

	err = hc.ghostwriter.database.QueryRow(`SELECT EXISTS(SELECT 1 FROM Tasks WHERE uuid = ? AND task = ?)`, uuid, task).Scan(&exists)
	if err != nil {
		hc.LogDbgError("db.QueryRow failed: %v", err)
		return false, err
	}

	return exists, nil
}

func (hc *GhostWriterPlugin) TaskQueryDescription(uuid string, task int) (string, error) {
	var (
		stmt   *sql.Stmt
		err    error
		exists bool
		value  string
	)

	if exists, err = hc.TaskExists(uuid, task); err != nil {
		return value, err
	} else if !exists {
		return value, errors.New("key does not exist")
	}

	stmt, err = hc.ghostwriter.database.Prepare("SELECT description FROM Tasks WHERE uuid = ? AND task = ?")
	if err != nil {
		hc.LogDbgError("sqlite.Prepare failed: %v", err)
		return "", err
	}

	err = stmt.QueryRow(uuid, task).Scan(&value)
	if err != nil {
		hc.LogDbgError("stmt.QueryRow failed: %v", err)
		return "", err
	}

	return value, err
}
