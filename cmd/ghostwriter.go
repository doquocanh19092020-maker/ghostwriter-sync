package main

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/k3a/html2text"

	havoc "github.com/InfinityCurveLabs/havoc-sdk/backend"
)

func (hc *GhostWriterPlugin) GhostWriter(event, uuid string, data map[string]any) {
	if event == havoc.TraceAgentEventConsole {
		if data["type"].(string) == "input" {
			hc.ghostwriter.executionInput[data["execution-uuid"].(string)] = data["text"].(string)
		} else if data["type"].(string) == "text" {
			var (
				err   error
				_text string
				_html bool
			)

			if MapExists(data, "is-html") {
				if _html, err = MapKey[bool](data, "is-html"); err != nil {
					return
				}
			}

			if _text, err = MapKey[string](data, "text"); err != nil {
				return
			}

			if _html {
				_text = html2text.HTML2Text(_text)
				_text = strings.ReplaceAll(_text, "&nbsp;", " ") // we are replacing some html "spaces" with real ones
			}

			if MapExists(data, "execution-uuid") {
				hc.ghostwriter.client.UpdateLogOutput(data["execution-uuid"].(string), _text)
			}
		}
	}

	if event == havoc.TraceAgentEventTask {
		var (
			err    error
			agent  map[string]any
			parent map[string]any
			puuid  string
			task   int
			source string
			dest   string
			entry  LogEntry
		)

		if !MapExists(data, "task-uuid") {
			return
		}

		task, err = MapKey[int](data, "task-uuid")
		if err != nil {
			hc.LogDbgError("failed to get task uuid: %v", err)
			return
		}

		agent, err = hc.AgentMetadata(uuid)
		if err != nil {
			hc.LogDbgError("failed to get agent metadata: %v", err)
			return
		}

		if puuid, err = hc.DatabaseAgentParent(uuid); err == nil && len(puuid) > 0 {
			parent, err = hc.AgentMetadata(puuid)
			if err != nil {
				hc.LogDbgError("failed to get parent agent metadata: %v", err)
				return
			}

			source = fmt.Sprintf("%v@%v (%v)", parent["user"], parent["domain"], parent["local ip"])
		}

		dest = fmt.Sprintf("%v@%v (%v)", agent["user"], agent["domain"], agent["local ip"])

		entry = LogEntry{
			StartTime:   time.Now(),
			EndTime:     time.Now(),
			SourceIP:    source,
			DestIP:      dest,
			Tool:        "Havoc Kaine",
			UserContext: fmt.Sprintf("%v\\%v", agent["host"], agent["user"]),
			Comments:    fmt.Sprintf("Task %x", task),
			Operator:    data["user"].(string),
			ExtraFields: map[string]any{
				"task-uuid": task,
			},
		}

		if MapExists(data, "execution-uuid") {
			if val, ok := hc.ghostwriter.executionInput[data["execution-uuid"].(string)]; ok {
				entry.Command = val

				// we no longer need to keep track fo the execution input
				delete(hc.ghostwriter.executionInput, data["execution-uuid"].(string))
			}
		}

		entry.Description, _ = hc.TaskQueryDescription(uuid, int(task))

		err = hc.ghostwriter.client.SendLog(data["execution-uuid"].(string), entry)
		if err != nil {
			hc.LogError("failed to ghostwriter send log: %v", err)
		}
	}
}

func MapExists(m map[string]any, key string) bool {
	_, ok := m[key]

	return ok
}

func MapKey[T any](m map[string]any, key string) (T, error) {
	var (
		zero T
		ok   bool
		val  any
	)

	if val, ok = m[key]; !ok {
		return zero, fmt.Errorf("key %q not found", key)
	}

	if val == nil {
		return zero, nil
	}

	if expected := reflect.TypeOf(zero); expected != nil {
		if expected.Kind() == reflect.Int {
			if f, ok := val.(float64); ok {
				return any(int(f)).(T), nil
			}
		}
	}

	typed, ok := val.(T)
	if !ok {
		return zero, fmt.Errorf(
			"key %q is of type %T, expected type is %s",
			key,
			val,
			reflect.TypeOf((*T)(nil)).Elem(),
		)
	}

	return typed, nil
}
