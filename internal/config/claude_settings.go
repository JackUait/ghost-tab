package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MergeResult indicates what MergeStatusLine did.
type MergeResult int

const (
	// StatusLineCreated means a new file was created with the statusLine.
	StatusLineCreated MergeResult = iota
	// StatusLineAdded means statusLine was added to an existing file.
	StatusLineAdded
	// StatusLineExists means statusLine was already present — no changes made.
	StatusLineExists
)

func (r MergeResult) String() string {
	switch r {
	case StatusLineCreated:
		return "created"
	case StatusLineAdded:
		return "added"
	case StatusLineExists:
		return "exists"
	default:
		return "unknown"
	}
}

// HookResult indicates what AddSoundHook did.
type HookResult int

const (
	// HookAdded means the hook was added to the file.
	HookAdded HookResult = iota
	// HookExists means the hook was already present — no changes made.
	HookExists
)

func (r HookResult) String() string {
	switch r {
	case HookAdded:
		return "added"
	case HookExists:
		return "exists"
	default:
		return "unknown"
	}
}

// MergeStatusLine adds a statusLine configuration to the Claude settings file
// if one doesn't already exist. Creates the file if it doesn't exist.
// The statusLine parameter is the value to set for the "statusLine" key.
func MergeStatusLine(path string, statusLine map[string]interface{}) (MergeResult, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return 0, fmt.Errorf("creating parent directories: %w", err)
	}

	settings, fileExisted, err := readSettingsFile(path)
	if err != nil {
		return 0, err
	}

	// Check if statusLine already exists
	if _, ok := settings["statusLine"]; ok {
		return StatusLineExists, nil
	}

	// Add the statusLine
	settings["statusLine"] = statusLine

	if err := writeSettingsFile(path, settings); err != nil {
		return 0, err
	}

	if fileExisted {
		return StatusLineAdded, nil
	}
	return StatusLineCreated, nil
}

// AddSoundHook adds a sound notification hook to the Claude settings file
// if it doesn't already exist. Creates the file if it doesn't exist.
// The hookCommand is the shell command to run for the notification.
//
// The hook structure follows the Claude Code settings format:
//
//	{
//	  "hooks": {
//	    "Notification": [
//	      {
//	        "matcher": "idle_prompt",
//	        "hooks": [
//	          {
//	            "type": "command",
//	            "command": "<hookCommand>"
//	          }
//	        ]
//	      }
//	    ]
//	  }
//	}
func AddSoundHook(path string, hookCommand string) (HookResult, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return 0, fmt.Errorf("creating parent directories: %w", err)
	}

	settings, _, err := readSettingsFile(path)
	if err != nil {
		return 0, err
	}

	// Get or create "hooks" object
	hooksObj, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		hooksObj = make(map[string]interface{})
		settings["hooks"] = hooksObj
	}

	// Get or create "Notification" array
	var notifList []interface{}
	if existing, ok := hooksObj["Notification"].([]interface{}); ok {
		notifList = existing
	}

	// Check if idle_prompt hook already exists
	for _, item := range notifList {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if entry["matcher"] == "idle_prompt" {
			return HookExists, nil
		}
	}

	// Build the new hook entry
	newEntry := map[string]interface{}{
		"matcher": "idle_prompt",
		"hooks": []interface{}{
			map[string]interface{}{
				"type":    "command",
				"command": hookCommand,
			},
		},
	}

	notifList = append(notifList, newEntry)
	hooksObj["Notification"] = notifList

	if err := writeSettingsFile(path, settings); err != nil {
		return 0, err
	}

	return HookAdded, nil
}

// readSettingsFile reads and parses a JSON settings file.
// Returns the parsed map, whether the file existed, and any error.
// If the file doesn't exist or contains invalid JSON, returns an empty map.
// Returns an error only for permission/IO issues that prevent reading.
func readSettingsFile(path string) (map[string]interface{}, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), false, nil
		}
		return nil, false, fmt.Errorf("reading settings file: %w", err)
	}

	// Try to parse the JSON
	settings := make(map[string]interface{})
	content := strings.TrimSpace(string(data))
	if content == "" {
		return settings, true, nil
	}

	if err := json.Unmarshal([]byte(content), &settings); err != nil {
		// Malformed JSON: start fresh (matches bash/python behavior)
		return make(map[string]interface{}), true, nil
	}

	return settings, true, nil
}

// writeSettingsFile marshals the settings map and writes it to the file
// with 2-space indentation and a trailing newline.
func writeSettingsFile(path string, settings map[string]interface{}) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}

	// Append trailing newline (matches python json.dump behavior)
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing settings file: %w", err)
	}

	return nil
}
